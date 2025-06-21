package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleEcho(update tgbotapi.Update) {
	if update.Message == nil || update.Message.From == nil || update.Message.Chat == nil {
		return
	}

	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	// Handle different message types
	switch {
	case update.Message.Text != "":
		b.echoTextMessage(chatID, userID, update.Message)
	case update.Message.Sticker != nil:
		b.echoStickerMessage(chatID, userID, update.Message)
	case update.Message.Photo != nil:
		b.echoPhotoMessage(chatID, userID, update.Message)
	case update.Message.Document != nil:
		b.echoDocumentMessage(chatID, userID, update.Message)
	case update.Message.Voice != nil:
		b.echoVoiceMessage(chatID, userID, update.Message)
	case update.Message.Video != nil:
		b.echoVideoMessage(chatID, userID, update.Message)
	case update.Message.Animation != nil:
		b.echoAnimationMessage(chatID, userID, update.Message)
	case update.Message.Audio != nil:
		b.echoAudioMessage(chatID, userID, update.Message)
	default:
		// Unsupported message type
		msg := tgbotapi.NewMessage(chatID, "🤖 Получил ваше сообщение, но не могу его повторить. Поддерживаются: текст, стикеры, фото, документы, голосовые, видео и аудио.")
		if _, err := b.api.Send(msg); err != nil {
			log.Printf("Failed to send unsupported message type response: %v", err)
		}
		log.Printf("Unsupported message type from user %d", userID)
	}
}

func (b *Bot) echoTextMessage(chatID int64, userID int64, message *tgbotapi.Message) {
	// Echo the text back with a prefix
	responseText := "📝 Вы написали: " + message.Text
	
	msg := tgbotapi.NewMessage(chatID, responseText)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Failed to echo text message: %v", err)
		return
	}
	
	log.Printf("Echoed text message from user %d: %s", userID, message.Text)
}

func (b *Bot) echoStickerMessage(chatID int64, userID int64, message *tgbotapi.Message) {
	// Send the same sticker back
	sticker := tgbotapi.NewSticker(chatID, tgbotapi.FileID(message.Sticker.FileID))
	if _, err := b.api.Send(sticker); err != nil {
		log.Printf("Failed to echo sticker: %v", err)
		return
	}
	
	log.Printf("Echoed sticker from user %d: %s", userID, message.Sticker.FileID)
}

func (b *Bot) echoPhotoMessage(chatID int64, userID int64, message *tgbotapi.Message) {
	// Get the largest photo size
	photo := message.Photo[len(message.Photo)-1]
	
	photoMsg := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(photo.FileID))
	if message.Caption != "" {
		photoMsg.Caption = "📸 Ваше фото: " + message.Caption
	} else {
		photoMsg.Caption = "📸 Получил ваше фото!"
	}
	
	if _, err := b.api.Send(photoMsg); err != nil {
		log.Printf("Failed to echo photo: %v", err)
		return
	}
	
	log.Printf("Echoed photo from user %d: %s", userID, photo.FileID)
}

func (b *Bot) echoDocumentMessage(chatID int64, userID int64, message *tgbotapi.Message) {
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileID(message.Document.FileID))
	if message.Caption != "" {
		doc.Caption = "📄 Ваш документ: " + message.Caption
	} else {
		doc.Caption = "📄 Получил ваш документ: " + message.Document.FileName
	}
	
	if _, err := b.api.Send(doc); err != nil {
		log.Printf("Failed to echo document: %v", err)
		return
	}
	
	log.Printf("Echoed document from user %d: %s", userID, message.Document.FileID)
}

func (b *Bot) echoVoiceMessage(chatID int64, userID int64, message *tgbotapi.Message) {
	voice := tgbotapi.NewVoice(chatID, tgbotapi.FileID(message.Voice.FileID))
	voice.Caption = "🎤 Ваше голосовое сообщение"
	
	if _, err := b.api.Send(voice); err != nil {
		log.Printf("Failed to echo voice: %v", err)
		return
	}
	
	log.Printf("Echoed voice message from user %d: %s", userID, message.Voice.FileID)
}

func (b *Bot) echoVideoMessage(chatID int64, userID int64, message *tgbotapi.Message) {
	video := tgbotapi.NewVideo(chatID, tgbotapi.FileID(message.Video.FileID))
	if message.Caption != "" {
		video.Caption = "🎥 Ваше видео: " + message.Caption
	} else {
		video.Caption = "🎥 Получил ваше видео!"
	}
	
	if _, err := b.api.Send(video); err != nil {
		log.Printf("Failed to echo video: %v", err)
		return
	}
	
	log.Printf("Echoed video from user %d: %s", userID, message.Video.FileID)
}

func (b *Bot) echoAnimationMessage(chatID int64, userID int64, message *tgbotapi.Message) {
	animation := tgbotapi.NewAnimation(chatID, tgbotapi.FileID(message.Animation.FileID))
	if message.Caption != "" {
		animation.Caption = "🎞️ Ваша GIF-анимация: " + message.Caption
	} else {
		animation.Caption = "🎞️ Получил вашу GIF-анимацию!"
	}
	
	if _, err := b.api.Send(animation); err != nil {
		log.Printf("Failed to echo animation: %v", err)
		return
	}
	
	log.Printf("Echoed animation from user %d: %s", userID, message.Animation.FileID)
}

func (b *Bot) echoAudioMessage(chatID int64, userID int64, message *tgbotapi.Message) {
	audio := tgbotapi.NewAudio(chatID, tgbotapi.FileID(message.Audio.FileID))
	if message.Caption != "" {
		audio.Caption = "🎵 Ваш аудиофайл: " + message.Caption
	} else {
		audio.Caption = "🎵 Получил ваш аудиофайл!"
	}
	
	if _, err := b.api.Send(audio); err != nil {
		log.Printf("Failed to echo audio: %v", err)
		return
	}
	
	log.Printf("Echoed audio from user %d: %s", userID, message.Audio.FileID)
}