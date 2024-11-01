package NomiKin

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "io/ioutil"
    "net/http"
)

var url = "https://api.kindroid.ai/v1/send-message" 

func (kin *NomiKin) SendKindroidMessage(message *string) (string, error) {
    if len(*message) > 749 {
        log.Printf("Message too long: %d", len(*message))
        return fmt.Sprintf("Your message was `%d` characters long, but the maximum message length is 750. Please send a shorter message.", len(*message)), nil
    }

    headers := map[string]string{
        "Authorization": "Bearer " + kin.ApiKey,
        "Content-Type": "application/json",
    }

    bodyMap := map[string]string{
        "message": *message,
        "ai_id": kin.CompanionId,
    }
    jsonBody, err := json.Marshal(bodyMap)
    jsonString := string(jsonBody)
    log.Printf("Sending message to Kin %v: %v", kin.CompanionId, jsonString)

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
       return "", fmt.Errorf("Error reading HTTP request: %v", err)
    }

    req.Header.Set("Authorization", headers["Authorization"])
    req.Header.Set("Content-Type", headers["Content-Type"])

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("Error sending HTTP request: %v", err)
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("Error reading HTTP response: %v", err)
    }

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("Error response from Kin API: %v", string(body))
    }

    kinReply := string(body)
    log.Printf("Received reply message from Kin %v: %v", kin.CompanionId, kinReply)
    return kinReply, nil
}
