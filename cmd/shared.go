package cmd

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/shahradelahi/cloudflare-warp/cloudflare/model"
)

func FormatMessage(shortMessage string, longMessage string) string {
	if longMessage != "" {
		longMessage = strings.TrimPrefix(longMessage, "\n")
		longMessage = strings.ReplaceAll(longMessage, "\n", " ")

	}
	if shortMessage != "" && longMessage != "" {
		return shortMessage + ". " + longMessage
	} else if shortMessage != "" {
		return shortMessage
	} else {
		return longMessage
	}
}

func F32ToHumanReadable(number float32) string {
	for i := 8; i >= 0; i-- {
		humanReadable := number / float32(math.Pow(1024, float64(i)))
		if humanReadable >= 1 && humanReadable < 1024 {
			return fmt.Sprintf("%.2f %ciB", humanReadable, "KMGTPEZY"[i-1])
		}
	}
	return fmt.Sprintf("%.2f B", number)
}

func PrintDeviceData(thisDevice *model.Identity, boundDevice *model.IdentityDevice) {
	log.Println("=======================================")
	log.Printf("% -13s : %s\n", "Device name", boundDevice.Name)
	log.Printf("% -13s : %s\n", "Device model", thisDevice.Model)
	log.Printf("% -13s : %t\n", "Device active", boundDevice.Active)
	log.Printf("% -13s : %s\n", "Account type", thisDevice.Account.AccountType)
	log.Printf("% -13s : %s\n", "Role", thisDevice.Account.Role)
	log.Printf("% -13s : %s\n", "Premium data", F32ToHumanReadable(float32(thisDevice.Account.PremiumData)))
	log.Printf("% -13s : %s\n", "Quota", F32ToHumanReadable(float32(thisDevice.Account.Quota)))
	log.Println("=======================================")
}
