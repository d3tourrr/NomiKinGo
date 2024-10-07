package NomiKin

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "io/ioutil"
    "net/http"
)

var urlComponents = []string {"https://api.nomi.ai/v1/nomis/", "/chat"}

func (nomi *NomiKin) SendNomiMessage (message *string) (string, error) {
    headers := map[string]string{
        "Authorization": nomi.ApiKey,
        "Content-Type": "application/json",
    }

    bodyMap := map[string]string{
        "messageText": *message,
    }
    jsonBody, err := json.Marshal(bodyMap)
    jsonString := string(jsonBody)
    log.Printf("Sending message to Nomi %v: %v", nomi.CompanionId, jsonString)

    url := urlComponents[0] + nomi.CompanionId + urlComponents[1]
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return "", fmt.Errorf("Error reading HTTP request: %v", err)
    }

    req.Header.Set("Authorization", headers["Authorization"])
    req.Header.Set("Content-Type", headers["Content-Type"])

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("Error making HTTP request: %v", err)
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("Error reading HTTP response: %v", err)
    }

    if resp.StatusCode != http.StatusOK {
        returnMsg := ""
        // Sometimes Nomi responds with an error: {"error":{"type":"NoReply"}}
        // The Nomi got the message and replied, but the reply wasn't sent back
        // This is only one kind of error, but it seems to be common enough that we
        // Should send it back as a message and let them know something happened
        var errorResult map[string]interface{}
        if err := json.Unmarshal(body, &errorResult); err != nil {
            return "", fmt.Errorf("Error unmarshalling error response: %v", err)
        }

        if errorMessage, ok := errorResult["error"].(map[string]interface{}); ok {
            if typeValue, ok := errorMessage["type"].(string); ok {
                if typeValue == "NoReply" {
                    log.Print("'NoReply' error - Sending 'Replied but you did not see it' message")
                    // Send as a reply to the message that triggered the response, helps keep things orderly
                    returnMsg = "❌ ERROR! ❌\nI got your message and I replied to it, but I took too long to do it so the Nomi API timed out. My response is available in the Nomi app and I have no idea that you didn't get it here. Try saying 'Sorry, I missed what you said when I said...' and send me your message again."
                }
            }
        }
        return returnMsg, fmt.Errorf("Error response from Nomi API: %v", string(body))
    }

    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        return "", fmt.Errorf("Error unmarshalling Nomi %v response: %v", nomi.CompanionId, err)
    }

    if replyMessage, ok := result["replyMessage"].(map[string]interface{}); ok {
        log.Printf("Received reply message from Nomi %v: %v\n", nomi.CompanionId, replyMessage)
        if textValue, ok := replyMessage["text"].(string); ok {
            return textValue, nil
        }
    }

    return "", fmt.Errorf("Failed to return anything meaningful")
}
