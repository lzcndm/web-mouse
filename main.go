package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/go-vgo/robotgo"

	"github.com/gorilla/websocket"
)

const (
	TypeMove = iota
	TypeClick
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type MoveTo struct {
	Type int
	X    int
	Y    int
}

var MoveToChan = make(chan MoveTo, 100)

func serveIndex(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tpl.Execute(w, "ws://"+serverIP+"/server")
}

func serveMoveTo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)

		println(message[0])
		var moveTo MoveTo
		err = json.Unmarshal(message, &moveTo)
		if err == nil {
			fmt.Printf("%v\n", moveTo)
			MoveToChan <- moveTo
		} else {
			println(err.Error())
		}
	}
}

func MoveMouse() {
	x, y := robotgo.GetMousePos()
	println("x:", x, "y:", y)
	for {
		moveTo := <-MoveToChan
		switch moveTo.Type {
		case TypeMove:
			x, y = x+moveTo.X, y+moveTo.Y
			println("move to x:", x, "y:", y)
			robotgo.Move(x, y)
		case TypeClick:
			println("click")
			println("click to x:", x, "y:", y)
			robotgo.MouseClick("left", true)
		default:
			println("???")
		}
	}
}

func main() {
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/server", serveMoveTo)
	go MoveMouse()
	log.Fatal(http.ListenAndServe(":8091", nil))
}

func getIP() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}

		}
	}
	return ""
}

var serverIP = getIP()

var tpl = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Document</title>
</head>
<style>
    #body {
        margin: 0;
        height: 100vh;
    }
    #poi {
        width: 80vw;
        height: 80vh;
        margin: 0 auto;
        background-color: yellow;
    }
</style>
<body id="body">
    <div id="poi">

    </div>
</body>
<script>
    let body = document.getElementById("poi")
    let preX = preY = 0
    let ws = new WebSocket("{{.}}")
    body.addEventListener('touchstart', (e) => {
        let touch = e.touches[0]
        preX = touch.pageX
        preY = touch.pageY
        console.log('start')
    },false)
    body.addEventListener('touchmove', (e) => {
        e.preventDefault()
        let touch = e.touches[0]
        let trendX = touch.pageX - preX
        let trendY = touch.pageY - preY
        preX = touch.pageX
        preY = touch.pageY
        ws.send(JSON.stringify({
            x: Number(trendX.toFixed()),
            y: Number(trendY.toFixed())
        }))
    },false)
    body.addEventListener('touchend', (e) => {
        let touch = e.touches[0]
        console.log('end',touch)
    },false)
    body.addEventListener('click', () => {
        ws.send(JSON.stringify({
            type: 1
        }))
    }, false)
    body.addEventListener('dbclick', () => {
        console.log('dbclick')
    }, false)
</script>
</html>
`))
