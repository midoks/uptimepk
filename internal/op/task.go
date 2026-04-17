package op

import (
	"fmt"

	"uptimepk/internal/db"
	"uptimepk/internal/tgtask"
)

func InitTgTask() {
	manager := tgtask.GetManager()

	recipient_data, err := db.GetAdminRecipientsInstancesByID(1)
	if err != nil {
		fmt.Printf("Failed to get recipient data: %v\n", err)
		return
	}

	if recipient_data == nil {
		fmt.Println("No recipient data found")
		return
	}

	fmt.Printf("Recipient data: ID=%d, MediaType=%s\n", recipient_data.ID, recipient_data.MediaType)

	if recipient_data.MediaType == "telegram" {
		tp, err := recipient_data.GetTelegramParams()
		if err != nil {
			fmt.Printf("Failed to get telegram params: %v\n", err)
			return
		}

		fmt.Printf("Telegram params: Token=%s, Proxy=%s\n", tp.Token, recipient_data.GetTelegramProxy())

		botID := recipient_data.ID
		if err := manager.AddBot(botID, tp.Token, recipient_data.GetTelegramProxy(), 0); err != nil {
			fmt.Printf("Failed to add bot: %v\n", err)
			return
		}

		if err := manager.StartBot(botID); err != nil {
			fmt.Printf("Failed to start bot: %v\n", err)
			return
		}

		fmt.Printf("Bot %d started successfully\n", botID)
	} else {
		fmt.Printf("MediaType is not telegram: %s\n", recipient_data.MediaType)
	}
}
