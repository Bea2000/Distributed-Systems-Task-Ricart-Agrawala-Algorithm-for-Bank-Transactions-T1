package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Request struct {
	// a request to enter the critical section
	turn int
	id   int
}

type Account struct {
	// an account in the bank
	id              int
	turn            int
	highestTurn     int
	requestCS       bool
	last_message_id int
	deferred_queue  []Request
	deferred_mutex  sync.Mutex
}

type Message struct {
	// a message sent between accounts only when they are in the critical section
	from  int
	money int
	to    int
	time  int
}

// a map of channels for sending requests to enter the critical section
var requestChannels = make(map[int]chan Request)

// a map of channels for sending approvals to enter the critical section
var approveChannels = make(map[int]chan int)

func createChannels(accounts []Account) {
	// create a channel for sending requests to enter the critical section
	for i := range accounts {
		requestChannels[i] = make(chan Request)
		approveChannels[i] = make(chan int)
	}
}

func NewAccount(id int) Account {
	// create a new account with the given id and money
	return Account{
		id:              id,
		turn:            0,
		highestTurn:     0,
		requestCS:       false,
		deferred_queue:  make([]Request, 0),
		last_message_id: 0,
	}
}

func (account *Account) NewRequest() Request {
	// create a new request with the given turn and id
	return Request{
		turn: account.turn,
		id:   account.id,
	}
}

func (account *Account) sendRequest(request Request, accounts []Account) {
	// send the request to all accounts except the one that made the request
	for i := 0; i < len(accounts); i++ {
		if i != account.id {
			requestChannels[i] <- request
		}
	}
}

func (account *Account) approveRequest(request Request) {
	// send an approval to the account that made the request
	approveChannels[request.id] <- account.id
}

func waitForApproval(account *Account, accounts []Account) {
	// wait for all accounts to approve the request
	for i := 0; i < len(accounts)-1; i++ {
		<-approveChannels[account.id]
	}
}

func (account *Account) askCS(request Request, accounts []Account) {
	// ask to enter the critical section
	account.turn += account.highestTurn + 1
	request.turn = account.turn
	account.requestCS = true
	account.sendRequest(request, accounts)
	waitForApproval(account, accounts)
}

func (account *Account) releaseCS() {
	// release the critical section
	account.requestCS = false
	account.deferred_mutex.Lock()
	for len(account.deferred_queue) > 0 {
		request := account.deferred_queue[0]
		account.deferred_queue = account.deferred_queue[1:]
		account.approveRequest(request)
	}
	account.deferred_mutex.Unlock()
}

func (account *Account) receiveRequest(request Request) {
	// receive a request to enter the critical section
	// change highetsTurn to the highest turn received
	if request.turn > account.highestTurn {
		account.highestTurn = request.turn
	}
	if !account.requestCS || (request.turn < account.turn) || (request.turn == account.turn && request.id < account.id) {
		account.approveRequest(request)
	} else {
		account.deferred_mutex.Lock()
		account.deferred_queue = append(account.deferred_queue, request)
		account.deferred_mutex.Unlock()
	}
}

func registerFinalBalances(accounts []Account) {
	// create a file to write the final balances of the accounts
	file, err := os.Create("saldo.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	for i := 0; i < len(accounts); i++ {
		total_money := checkAvailableMoney(i)
		file.WriteString(fmt.Sprintf("%d,%d\n", i, total_money))
	}
}

func registerTransaction(message Message) {

	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println("error al abrir el archivo de transacciones:", err)
		return
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("Participante %d ha transferido %d a participante %d.\n", message.from, message.money, message.to))
}

func checkAvailableMoney(id int) int {
	var final_money int = 0
	file, err := os.Open("logs.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	// check all transactions that involve the account with the given id
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if len(parts) < 8 {
			fmt.Println("Formato de línea incorrecto:", line)
			continue
		}

		from_str := strings.TrimPrefix(parts[1], "Participante")
		from, err := strconv.Atoi(from_str)
		if err != nil {
			fmt.Println("Error al parsear participante 1:", err)
			continue
		}

		money_str := strings.TrimPrefix(parts[4], "$")
		money, err := strconv.Atoi(money_str)
		if err != nil {
			fmt.Println("Error al parsear el dinero:", err)
			continue
		}

		to_str := strings.TrimPrefix(parts[7], "participante")
		to_str = strings.TrimSuffix(to_str, ".")
		to, err := strconv.Atoi(to_str)
		if err != nil {
			fmt.Println("Error al parsear participante 2:", err)
			continue
		}

		if from == id {
			final_money -= money
		} else if to == id {
			final_money += money
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error leyendo el archivo:", err)
	}

	return final_money
}

func readTransactions(folder_name string) ([]Account, []Message) {
	// Abrir el archivo
	file, err := os.Open(folder_name + "/transacciones.txt")
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	defer file.Close()

	// Crear un scanner para leer el archivo línea por línea
	scanner := bufio.NewScanner(file)

	// Leer la primera línea que contiene el número de cuentas y transacciones
	scanner.Scan()
	firstLine := scanner.Text()
	parts := strings.Split(firstLine, ",")
	n_accounts, _ := strconv.Atoi(parts[0])
	m_transactions, _ := strconv.Atoi(parts[1])

	// Crear el array de cuentas
	var accounts = make([]Account, n_accounts)
	for i := 0; i < n_accounts; i++ {
		accounts[i] = NewAccount(i)
	}

	// Crear el array de transacciones (mensajes)
	var messages = make([]Message, m_transactions)

	// Leer el resto de las líneas que contienen las transacciones
	i := 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		from, _ := strconv.Atoi(parts[0])
		money, _ := strconv.Atoi(parts[1])
		to, _ := strconv.Atoi(parts[2])
		time, _ := strconv.Atoi(parts[3])

		messages[i] = Message{
			from:  from,
			to:    to,
			money: money,
			time:  time,
		}

		i++
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error leyendo el archivo:", err)
	}

	return accounts, messages
}

func (account *Account) processTransaction(messages []Message, accounts []Account, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := account.last_message_id + 1; i < len(messages); i++ {
		message := messages[i]
		if message.from == account.id {
			account.last_message_id = i
			account.askCS(account.NewRequest(), accounts)

			if checkAvailableMoney(account.id) < message.money {
				account.releaseCS()
				for checkAvailableMoney(account.id) < message.money {
				}
				account.askCS(account.NewRequest(), accounts)
			}

			registerTransaction(message)
			account.releaseCS()

			if message.time > 0 {
				time.Sleep(time.Duration(message.time) * time.Millisecond)
			}
		}

	}
}

func main() {

	os.Remove("logs.txt")
	folder_name := "tests/test_10"

	accounts, messages := readTransactions(folder_name)

	// create channels for sending requests, approvals, and messages
	createChannels(accounts)

	// create a goroutine for each account for process receive requests
	for i := range accounts {
		go func(account *Account) {
			for request := range requestChannels[account.id] {
				account.receiveRequest(request)
			}
		}(&accounts[i])
	}

	// process bank transactions
	for i := range accounts {
		registerTransaction(messages[i])
	}

	// create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// create a goroutine for each account for process messages
	for i := range accounts {
		wg.Add(1)
		go func(account *Account) {
			account.processTransaction(messages, accounts, &wg)
		}(&accounts[i])
	}

	// wait for all goroutines to finish
	wg.Wait()

	// register the final balances of the accounts
	registerFinalBalances(accounts)
}
