package main
           
import (
        "fmt"
        "net/smtp"
	"log"
	"os"
	"time"
	"strings"
	"io/ioutil"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/disk"
)          
           
func main() {
	targetUser := "2693389741@msg.fi.google.com"
	content, err := ioutil.ReadFile("/home/ubuntu/.sysMgrCreds")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	tmp := strings.Split(string(content), ":")
	emailUser := tmp[0]
	emailPass := tmp[1]

	sendFlag := false
	hostname, _ := os.Hostname()
	emailContent := fmt.Sprintf("\nHost: %s\n", hostname)
        //check CPU, MEM, DISK. Alert if over 80% anywhere
	percent, err := cpu.Percent(1 * time.Second, false)
	if err != nil {
		log.Fatal(err)
	}
	if percent[0] > 80.0 {
		fmt.Printf("High CPU Usage: %.2f%%\n", percent[0])
		emailContent += fmt.Sprintf("\nHigh CPU Usage: %.2f%%\n", percent[0])
		sendFlag = true
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal(err)
	}
	if vm.UsedPercent > 80.0 {
		fmt.Printf("High Memory Usage: %.2f%%\n", vm.UsedPercent)
		emailContent += fmt.Sprintf("\nHigh Memory Usage: %.2f%%\n", vm.UsedPercent)
		sendFlag = true
	}

	diskStat, err := disk.Usage("/")
	if err != nil {
		log.Fatal(err)
	}
	if diskStat.UsedPercent > 80.0 {
		emailContent += fmt.Sprintf("\nHigh Disk Usage: %.2f%%\n", diskStat.UsedPercent)
		sendFlag = true
	}

	if sendFlag {
		//connect to our server, set up a message and send it
		auth := smtp.PlainAuth("", emailUser, emailPass, "smtp.gmail.com")
		   
		mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		from := emailUser
		to := []string{targetUser}
		msg := []byte(fmt.Sprintf("To: %s\r\n", targetUser) +
			"Subject: EW SYSTEM CRITICAL\r\n" +
			mime +
			emailContent)
		   
		err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)
		if err != nil {
			fmt.Println("unable to send email to gmail server: ", err)
		}  
        }
}          
     

