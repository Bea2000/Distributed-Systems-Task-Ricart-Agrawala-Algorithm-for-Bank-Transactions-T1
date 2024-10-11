# Distributed Systems Task: Ricart-Agrawala Algorithm for Bank Transactions

This project is an implementation of the Ricart-Agrawala algorithm to handle critical section problems related to distributed bank transactions. The program simulates multiple bank accounts that perform concurrent transactions, ensuring consistency using distributed mutual exclusion. This was made for the course of Distributed Systems at the Pontifica Catholic University on the first semester of 2024 as a Engineering Computer Science student.

## Features

- **Concurrency**: Handles transactions concurrently between multiple bank accounts.
- **Critical Section Management**: Ensures only one account can process a transaction at a time using the Ricart-Agrawala algorithm.
- **Transaction Logging**: Records every successful transaction in a log file (`logs.txt`).
- **Final Balances**: After all transactions are completed, the final balance of each account is written to `saldo.txt`.
- **Fault Tolerance**: Automatically handles cases where accounts attempt to transfer more money than they have by waiting for other participants to transfer sufficient funds.

## Requirements

- **Go**: Version 1.21 or higher.

## Running the Project

1. Clone the Repository

    ```bash
    git clone https://github.com/yourusername/repo.git
    cd repo
    ```

2. Edit the tests file on `main.go` modifying the `folder_name` variable on line **295**.

    ```go
    folder_name := "tests/test_XX"
    ```

3. Build and Run the Project

    ```bash
    go run main.go
    ```

This will run the program and generate two output files:

- `logs.txt`: A log of every transaction performed.
- `saldo.txt`: Final balance of each account.

## Example Transactions File

The transactions file should be a `transaccciones.txt` file with the following format:

```txt
n_accounts n_transactions
account_number1 amount account_number2 sleep_time
account_number1 amount account_number2 sleep_time
...
```

## How It Works

- **Initialization**: The program reads from transacciones.txt, initializing n accounts and recording m transactions. Each transaction consists of a sender, an amount of money, a recipient, and the time it takes to process.
- **Concurrent Processing**: Each account processes its own transactions concurrently. Before performing a transaction, the account requests access to the critical section.
- **Critical Section**: Once an account enters the critical section, it verifies its balance by reading the history of previous transactions in logs.txt. If the balance is sufficient, it proceeds with the transaction and records it.
- **Transaction Recording**: The program records every transaction in logs.txt and, after all transactions are completed, the final balances are written to saldo.txt.

## Output

1. `logs.txt`: Records all successful transactions in the format:

    ```txt
    Participant {from_account} has transferred ${amount} to participant {to_account}.
    ```

2. `saldo.txt`: Contains the final balance of each account in the format:

    ```txt
    account_id,final_balance
    ```
