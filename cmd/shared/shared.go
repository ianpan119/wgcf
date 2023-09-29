package shared

import (
	"fmt"
	"log"
	"math"
	"strings"
	"strconv"

	"encoding/base64"
	"encoding/hex"
	
	"github.com/ViRb3/wgcf/cloudflare"
	"github.com/ViRb3/wgcf/config"
	"github.com/ViRb3/wgcf/util"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func FormatMessage(shortMessage string, longMessage string) string {
	if longMessage != "" {
		if strings.HasPrefix(longMessage, "\n") {
			longMessage = longMessage[1:]
		}
		longMessage = strings.Replace(longMessage, "\n", " ", -1)
	}
	if shortMessage != "" && longMessage != "" {
		return shortMessage + ". " + longMessage
	} else if shortMessage != "" {
		return shortMessage
	} else {
		return longMessage
	}
}

func IsConfigValidAccount() bool {
	return viper.GetString(config.DeviceId) != "" &&
		viper.GetString(config.AccessToken) != "" &&
		viper.GetString(config.PrivateKey) != ""
}

func CreateContext() *config.Context {
	ctx := config.Context{
		DeviceId:    viper.GetString(config.DeviceId),
		AccessToken: viper.GetString(config.AccessToken),
		LicenseKey:  viper.GetString(config.LicenseKey),
	}
	return &ctx
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

func PrintDeviceData(thisDevice *cloudflare.Device, boundDevice *cloudflare.BoundDevice) {
	clientID := thisDevice.Config.ClientId

	// 解碼base64編碼的字串將其轉換為十六進制
	decoded, err := base64.StdEncoding.DecodeString(clientID)
	if err != nil {
		fmt.Println(err)
		return
	}
	hexString := hex.EncodeToString(decoded)

	// 將十六行字串轉換為十行值並列印它們
	var decValues []string
	for i := 0; i < len(hexString); i += 2 {
		hexByte := hexString[i : i+2]
		decValue, _ := strconv.ParseInt(hexByte, 16, 64)
		decValues = append(decValues, fmt.Sprintf("%d%d%d", decValue/100, (decValue/10)%10, decValue%10))
	}

	reserved := []int{}
	for i := 0; i < len(hexString); i += 2 {
		hexByte := hexString[i : i+2]
		decValue, _ := strconv.ParseInt(hexByte, 16, 64)
		reserved = append(reserved, int(decValue))
	}

	// 使用一个字符串切片收集所有的整数字符串
	strValues := make([]string, len(reserved))
	for i, value := range reserved {
		strValues[i] = strconv.Itoa(value)
	}
	
	log.Println("=======================================")
	log.Printf("%-13s : %s\n", "Device name", *boundDevice.Name)
	log.Printf("%-13s : %s\n", "Device model", thisDevice.Model)
	log.Printf("%-13s : %t\n", "Device active", boundDevice.Active)
	log.Printf("%-13s : %s\n", "Account type", thisDevice.Account.AccountType)
	log.Printf("%-13s : %s\n", "Role", thisDevice.Account.Role)
	log.Printf("%-13s : %s\n", "Premium data", F32ToHumanReadable(thisDevice.Account.PremiumData))
	log.Printf("%-13s : %s\n", "Quota", F32ToHumanReadable(thisDevice.Account.Quota))
	log.Printf("%-13s : %s\n", "Client ID", thisDevice.Config.ClientId)
	log.Printf("%-13s : %s\n", "Reserved", strings.Join(strValues, ", "))
	log.Println("=======================================")
}

// changing the bound account (e.g. changing license key) will reset the device name
func SetDeviceName(ctx *config.Context, deviceName string) (*cloudflare.BoundDevice, error) {
	if deviceName == "" {
		deviceName += util.RandomHexString(3)
	}
	device, err := cloudflare.UpdateSourceBoundDeviceName(ctx, deviceName)
	if err != nil {
		return nil, err
	}
	if device.Name == nil || *device.Name != deviceName {
		return nil, errors.New("could not update device name")
	}
	return device, nil
}