# LogViz

This Go package offers a request logging and viewing system designed for web applications built with frameworks like Chi or Gin. It enables developers to log and review HTTP requests sent to their applications, aiding in debugging and monitoring processes.

![Logviz](https://res.cloudinary.com/rammy/image/upload/v1711575024/logwiz.png)

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Installation

Run the following command to install Horus on your project:

```bash
go get github.com/rammyblog/logviz
```

![logviz2](https://res.cloudinary.com/rammy/image/upload/v1711575024/logwiz2.png)


## Usage

```

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rammyblog/logviz"
)

// Example handler
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	time.Sleep(2 * time.Millisecond)
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte("Hello world"))
}

func SecondHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte("second handler"))
}

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	reqLogger, err := logviz.Init("postgres", logviz.DbConfig{
		DbUser:     "postgres",
		DbPassword: "password",
		DbHost:     "localhost",
		DbName:     "logviz",
		DbPort:     "5429",
	})

	if err != nil {
		fmt.Println(err)
	}
	reqLogger.Serve(":5009")

	mux.Handle("/", reqLogger.Logger(http.HandlerFunc(HelloHandler)))
	mux.Handle("/second", reqLogger.Logger(http.HandlerFunc(SecondHandler)))

	// Start the server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

```

Then visit the site on `http://localhost:<PORT>` to view the logs

It only supports Postgres and MySQL for now, but support for other databases will be added soon.

## Contributing:
Contributions are welcome! Feel free to open issues or pull requests for bug fixes, enhancements, or new features

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

