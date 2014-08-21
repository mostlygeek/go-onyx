package main

import (
    "fmt"
    "time"
    "net/http"
    "sync/atomic"
    "encoding/json"
    "runtime"
)

func init() {
    runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

    var count uint64 = 0
    var bytes uint64 = 0

    logChan := make(chan []byte, 25)

    HandlePing := func(resp http.ResponseWriter, req *http.Request) {

        if req.Method != "POST" {
            http.Error(resp, "POST Request expected", http.StatusBadRequest)
            return
        }

        decoder := json.NewDecoder(req.Body)

        var payload interface{}
        err := decoder.Decode(&payload)
        if err != nil {
            http.Error(resp, "Invalid JSON", http.StatusBadRequest)
            return
        }

        m := payload.(map[string]interface{})

        m["ip"] = req.RemoteAddr
        m["ua"] = req.Header.Get("User-Agent")

        now := time.Now()
        m["timestamp"] = int32(now.Unix())
        m["date"] = fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day())

        b, err := json.Marshal(m)
        if err != nil {
            http.Error(resp, "Error encoding JSON", http.StatusInternalServerError)
            return
        }

        // send it into the channel
        logChan <- b
    }

    go func() {
        for {
            /* log it out to disk here ... */
            //_ = <-logChan
            fmt.Println(string( <-logChan ))
        }
    }()


    fmt.Println("Listening on 9999")

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

        if r.ContentLength > 0 {
            atomic.AddUint64(&bytes, uint64(r.ContentLength))
        }

        atomic.AddUint64(&count, 1)
        fmt.Fprintf(w, "requests: %d, bytes: %d\n", atomic.LoadUint64(&count), atomic.LoadUint64(&bytes))

    })

    http.HandleFunc("/v2/links/view", HandlePing)
    http.HandleFunc("/v2/links/click", HandlePing)


    http.ListenAndServe(":9999", nil)
}

