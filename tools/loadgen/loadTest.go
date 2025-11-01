package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
)

func main() {
    baseURL := "http://localhost:8080"
    var wg sync.WaitGroup
    
    urls := []string{
        "https://example.com/page1",
        "https://example.com/page2", 
        "https://example.com/page3",
        "https://google.com/search",
        "https://github.com/golang",
    }
    
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            
            url := urls[idx%len(urls)]
            
            switch idx % 4 {
            case 0:
                // POST /api/shorten
                body := map[string]string{"url": url}
                jsonBody, _ := json.Marshal(body)
                resp, _ := http.Post(baseURL+"/api/shorten", "application/json", bytes.NewBuffer(jsonBody))
                if resp != nil && resp.Body != nil {
                    resp.Body.Close()
                }
                
            case 1:
                // POST /
                resp, _ := http.Post(baseURL+"/", "text/plain", bytes.NewBuffer([]byte(url)))
                if resp != nil && resp.Body != nil {
                    resp.Body.Close()
                }
                
            case 2:
                // GET /{id}
                if idx%10 == 0 {
                    resp, _ := http.Get(fmt.Sprintf("%s/nonexistent%d", baseURL, idx))
                    if resp != nil && resp.Body != nil {
                        resp.Body.Close()
                    }
                }
                
            case 3:
                // GET /api/user/urls
                req, _ := http.NewRequest("GET", baseURL+"/api/user/urls", nil)
                req.Header.Set("Authorization", "test-user")
                client := &http.Client{}
                resp, _ := client.Do(req)
                if resp != nil && resp.Body != nil {
                    resp.Body.Close()
                }
            }
        }(i)
        
        if i%100 == 0 {
            time.Sleep(10 * time.Millisecond)
        }
    }
    
    wg.Wait()
    fmt.Println("Load test completed")
}