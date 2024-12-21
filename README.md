# 🧮 Simple Go Calculator

Welcome to the **Simple Go Calculator** repository! This project provides a straightforward calculator implementation in Go, allowing you to perform basic arithmetic operations either in a **console mode** or via a **server mode**. Choose the mode that best suits your needs! \
[доступно на русском](./docs/READMERUS.md)

🚀 **Features**

**Console Mode**: Run the calculator directly in your terminal for quick calculations.

**Server Mode**: Start a lightweight HTTP server to handle calculations via API requests.
Supports basic arithmetic operations: `+`, `-`, `*`, `/`.
Easy to use and extend.

📦 **Installation**\
To get started, make sure you have Go installed on your machine. You can download it from [here](https://golang.org/dl/).

Clone the repository:

```bash
git clone https://github.com/shzuzu/Go_Calculator.git
cd Go_Calculator
```

🛠️ **Usage**

1. **Console Mode**
   Run the calculator in your terminal for interactive calculations:

```bash
go run ./cmd/main.go --mode=console
```

Enter your arithmetic expressions directly in the terminal.
**Example:**

```
Input expression (enter "exit" to exit):
2+2*2
2+2*2 = 6
```

2. **Server Mode**
   Start the calculator as an HTTP server to handle calculations via API:

```bash
go run ./cmd/main.go --mode=server
```

The server will start on `http://localhost:8080\`.
Send a POST request with your expression to `/api/v1/calculate`:

```bash
 curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
```

**Example response:**

```json
{
  "result": 6
}
```

🧪 **Testing**\
The project includes unit tests to ensure the calculator works as expected. Run the tests using:

```bash
go test ./internal/application
go test ./pkg/calc
```

📜 **License**\
This project is licensed under the **MIT License**. See the [LICENSE](./LICENSE) file for details.

💬 **Feedback**\
If you have any questions, suggestions, or issues, feel free to open an issue or reach out to me directly.

**Happy calculating!** 🎉

\
\
This is the end, fren..

BTW, THE CAT ↓

```

                   _ |\_
                   \` ..\
              __,.-" =__Y=
            ."        )
      _    /   ,    \/\_
     ((____|    )_-\ \_-`
     `-----'`-----` `--`

```
