# Wordpress Steadfast Order Status Update

Sync wordpress orders with steadfast order status.

### Getting started

- Clone repository
- Open terminal in the root folder of the project
- Execute `go mod tidy`
- Go to cmd folder in the terminal (ex. `cd cmd`)
- Execute `go build -o wpsfo.exe .`, and you will get the executable.

### Usage

- Copy the executable to a suitable folder
- Execute `.\wpsfo.exe --help`, you will have the available commands.

#### Available params

- `--host`: Wordpress host domain name
- `--consumer-key`: Wordpress consumer key
- `--consumer-secret`: Wordpress consumer secret
- `--steadfast-key`: Steadfast API key
- `--steadfast-secret`: Steadfast API secret
- `--statuses`: Order statuses that should check
- `--destinations`: Steadfast status map to order status map

#### Example

```
.\wpsfo.exe --host http://xyz.com --consumer-key WORDPRESS_CONSUMER_KEY --consumer-secret WORDPRESS_CONSUMER_SECRET --steadfast-key STEADFAST_API_KEY --steadfast-secret STEADFAST_API_SECRET --statuses pending,review --destinations delivered=delivered,partial_delivered=partially-deliver
```
