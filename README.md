# Ethereum Transaction Parser

A backend service in Go that tracks transactions for Ethereum addresses. It uses a RESTful API to subscribe to addresses and retrieve their transactions.


## Setup

1.  **Clone the repository:**

    ```bash
    git clone <repository_url>
    cd <project_directory>
    ```

2.  **Build and Run Docker Image:**

    Navigate to the root of your project where the `Dockerfile` is located and run:

    ```bash
    docker build . -t ehtxparser
    docker run -d -p <port>:<container_port> ethtxparser
    ```

    You can configure container port in the `config.json` file.

## API Usage

After starting the Docker environment, APIs should be accessible at `http://localhost:<port>`.

*  **Example**

    Send a POST request to `/api/address` to subscribe to an address.

    ```bash
    curl -X POST --location 'http://localhost:9600/api/subscribe?address=0x123'
    ```

    Send a GET request to `/api/transactions` to see incoming and outgoing transactions to an address.

    ```bash
    curl -X GET --location 'http://localhost:9600/api/transactions?address=0x123'
    ```

     Send a GET request to `/api/block` to get the last processed block number by the application.

    ```bash
    curl -X GET --location 'http://localhost:9600/api/block'
    ```