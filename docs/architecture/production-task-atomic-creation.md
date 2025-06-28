# Production Task Atomic Creation - Sequence Diagrams

## Overview

This document describes the sequence diagrams for atomic production task creation and inventory reservation using the Saga pattern.

## Current Implementation (Problematic)

```mermaid
sequenceDiagram
    participant Client as Frontend Client
    participant ProdSvc as Production Service
    participant DB as PostgreSQL
    participant InvSvc as Inventory Service
    
    Client->>ProdSvc: POST /production/factory/start
    
    Note over ProdSvc: Current Implementation (Non-Atomic)
    ProdSvc->>DB: BEGIN TRANSACTION
    ProdSvc->>DB: INSERT production_task (status=PENDING)
    ProdSvc->>DB: INSERT task_output_items
    ProdSvc->>DB: COMMIT
    
    ProdSvc->>InvSvc: POST /inventory/reserve
    alt Inventory Success
        InvSvc-->>ProdSvc: 200 OK
        ProdSvc-->>Client: 201 Task Created
    else Inventory Failure
        InvSvc-->>ProdSvc: 4xx/5xx Error
        Note over ProdSvc: Weak Compensation
        ProdSvc->>DB: UPDATE task status=FAILED
        ProdSvc-->>Client: 400/500 Error
        Note right of DB: 🚨 Task remains in DB!<br/>Inventory not reserved!
    end
```

**Problems:**
- Task exists in database even if inventory reservation fails
- No proper cleanup of failed tasks
- Data inconsistency between Production and Inventory services

## Proposed Solution: Saga Pattern

```mermaid
sequenceDiagram
    participant Client as Frontend Client
    participant ProdSvc as Production Service
    participant DB as PostgreSQL
    participant InvSvc as Inventory Service
    
    Client->>ProdSvc: POST /production/factory/start
    
    Note over ProdSvc: Phase 1: Draft Task Creation
    ProdSvc->>DB: BEGIN TRANSACTION
    ProdSvc->>DB: INSERT production_task (status=DRAFT)
    ProdSvc->>DB: INSERT task_output_items
    ProdSvc->>DB: COMMIT
    
    Note over ProdSvc: Phase 2: Inventory Reservation
    ProdSvc->>InvSvc: POST /inventory/reserve
    ProdSvc->>InvSvc: operationID: task.ID (idempotency)
    
    alt Reservation Success
        InvSvc-->>ProdSvc: 200 OK
        Note over ProdSvc: Phase 3: Task Confirmation
        ProdSvc->>DB: UPDATE production_task SET status=PENDING
        ProdSvc-->>Client: 201 Task Created Successfully
        
    else Reservation Failure
        InvSvc-->>ProdSvc: 4xx/5xx Error
        Note over ProdSvc: Compensation: Clean Rollback
        ProdSvc->>DB: DELETE production_task
        ProdSvc->>DB: DELETE task_output_items (CASCADE)
        ProdSvc-->>Client: 400/500 Error
        Note right of DB: ✅ Clean state!<br/>No orphaned data
    end
```

## Edge Case: Network Failure After Reservation

```mermaid
sequenceDiagram
    participant Client as Frontend Client
    participant ProdSvc as Production Service
    participant DB as PostgreSQL
    participant InvSvc as Inventory Service
    
    Client->>ProdSvc: POST /production/factory/start
    
    Note over ProdSvc: Phase 1: Draft Task Creation
    ProdSvc->>DB: INSERT task (status=DRAFT)
    
    Note over ProdSvc: Phase 2: Inventory Reservation
    ProdSvc->>InvSvc: POST /inventory/reserve
    InvSvc-->>ProdSvc: 200 OK (reservation successful)
    
    Note over ProdSvc: Phase 3: Confirmation
    ProdSvc->>DB: UPDATE status=PENDING
    Note over DB,ProdSvc: 💥 Network failure!<br/>Update fails
    
    Note over ProdSvc: Compensation Actions
    ProdSvc->>InvSvc: POST /inventory/return-reserve
    ProdSvc->>InvSvc: operationID: task.ID
    InvSvc-->>ProdSvc: 200 OK (reservation returned)
    
    ProdSvc->>DB: DELETE production_task
    ProdSvc-->>Client: 500 Internal Server Error
    
    Note right of DB: ✅ Consistent state restored<br/>No leaked reservations
```

## Background Cleanup Process

```mermaid
sequenceDiagram
    participant Scheduler as Background Scheduler
    participant ProdSvc as Production Service
    participant DB as PostgreSQL
    participant InvSvc as Inventory Service
    
    Note over Scheduler: Every 5 minutes
    Scheduler->>ProdSvc: Cleanup Orphaned Tasks
    
    ProdSvc->>DB: SELECT tasks WHERE status=DRAFT<br/>AND created_at < NOW() - INTERVAL '5 minutes'
    DB-->>ProdSvc: List of orphaned tasks
    
    loop For each orphaned task
        Note over ProdSvc: Check if inventory was reserved
        ProdSvc->>InvSvc: GET /inventory/reservation/{operationID}
        
        alt Reservation exists
            InvSvc-->>ProdSvc: 200 OK (reservation found)
            ProdSvc->>InvSvc: POST /inventory/return-reserve
            ProdSvc->>InvSvc: operationID: task.ID
            InvSvc-->>ProdSvc: 200 OK
        else No reservation
            InvSvc-->>ProdSvc: 404 Not Found
            Note over ProdSvc: Nothing to cleanup in inventory
        end
        
        ProdSvc->>DB: DELETE production_task
        Note over ProdSvc: Task cleaned up
    end
    
    Note over Scheduler: Cleanup completed
```

## Idempotency Handling

```mermaid
sequenceDiagram
    participant Client as Frontend Client
    participant ProdSvc as Production Service
    participant DB as PostgreSQL
    participant InvSvc as Inventory Service
    
    Client->>ProdSvc: POST /production/factory/start
    Note over Client: (Duplicate request due to network retry)
    
    Note over ProdSvc: Idempotency Check
    ProdSvc->>DB: SELECT * FROM production_tasks<br/>WHERE user_id = ? AND recipe_id = ?<br/>AND status IN ('DRAFT', 'PENDING')
    
    alt Existing task found
        DB-->>ProdSvc: Existing task data
        Note over ProdSvc: Return existing task
        ProdSvc-->>Client: 200 OK (existing task)
        
    else No existing task
        DB-->>ProdSvc: No results
        Note over ProdSvc: Proceed with normal flow
        ProdSvc->>DB: INSERT task (status=DRAFT)
        ProdSvc->>InvSvc: POST /inventory/reserve
        Note over InvSvc: operationID ensures<br/>idempotency in inventory
        InvSvc-->>ProdSvc: 200 OK
        ProdSvc->>DB: UPDATE status=PENDING
        ProdSvc-->>Client: 201 Created
    end
```

## Metrics and Monitoring

Key metrics to track the health of the atomic creation process:

- `production_saga_duration_seconds` - Total time for complete saga
- `production_saga_failures_total{phase="draft|reservation|confirmation"}`
- `production_compensation_actions_total{type="delete_task|return_reservation"}`
- `production_orphaned_tasks_cleaned_total`
- `production_idempotency_hits_total`

## Implementation Requirements

### Database Changes

```sql
-- Add DRAFT status to enum
ALTER TYPE task_status ADD VALUE 'DRAFT';

-- Index for cleanup operations
CREATE INDEX idx_production_tasks_draft_cleanup 
ON production_tasks(status, created_at) 
WHERE status = 'DRAFT';

-- Unique constraint for idempotency
CREATE UNIQUE INDEX idx_production_tasks_user_recipe_active
ON production_tasks(user_id, recipe_id)
WHERE status IN ('DRAFT', 'PENDING', 'IN_PROGRESS');
```

### Service Dependencies

- **Inventory Service** must support idempotent operations with `operationID`
- **Inventory Service** must provide `/inventory/return-reserve` endpoint
- **Production Service** needs background cleanup job
- **Monitoring** integration for saga metrics

This atomic creation pattern ensures **data consistency** and **fault tolerance** in distributed production task management.