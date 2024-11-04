package main

import (
	"fmt"
	"net/rpc"
	"os"
	"strconv"
	"sync"
)

type DepositoArgs struct {
	IDConta     int
	Valor       float64
	TransacaoID int
}

type SacarArgs struct {
	IDConta     int
	Valor       float64
	TransacaoID int
}

var transacaoCounter int
var transacaoMutex sync.Mutex

func gerarTransacaoID() int {
	transacaoMutex.Lock()
	defer transacaoMutex.Unlock()
	transacaoCounter++
	return transacaoCounter
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Uso:", os.Args[0], "<maquina> <operacao> [<id_conta> <valor>]")
		fmt.Println("Operações disponíveis: depositar, sacar, saldo")
		return
	}

	porta := 1234
	maquina := os.Args[1]
	operacao := os.Args[2]

	client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", maquina, porta))
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer client.Close()

	switch operacao {
	case "depositar":
		if len(os.Args) < 5 {
			fmt.Println("Uso para depositar:", os.Args[0], "<maquina> depositar <id_conta> <valor>")
			return
		}
		var resultado bool
		transacaoID := gerarTransacaoID()
		idConta := parseInt(os.Args[3])
		valor := parseFloat(os.Args[4])
		args := DepositoArgs{IDConta: idConta, Valor: valor, TransacaoID: transacaoID}
		err := client.Call("ServicoContas.Depositar", args, &resultado)
		if err != nil || !resultado {
			fmt.Printf("Erro ao depositar na conta ID %d: %v\n", idConta, err)
		} else {
			fmt.Printf("Depósito de %.2f realizado com sucesso na conta ID %d\n", valor, idConta)
		}

	case "sacar":
		if len(os.Args) < 5 {
			fmt.Println("Uso para sacar:", os.Args[0], "<maquina> sacar <id_conta> <valor>")
			return
		}
		var resultado bool
		transacaoID := gerarTransacaoID()
		idConta := parseInt(os.Args[3])
		valor := parseFloat(os.Args[4])
		args := SacarArgs{IDConta: idConta, Valor: valor, TransacaoID: transacaoID}
		err = client.Call("ServicoContas.Sacar", args, &resultado)
		if err != nil || !resultado {
			fmt.Printf("Erro ao sacar da conta ID %d: %v\n", idConta, err)
		} else {
			fmt.Printf("Saque de %.2f realizado com sucesso na conta ID %d\n", valor, idConta)
		}

	case "saldo":
		if len(os.Args) < 4 {
			fmt.Println("Uso para consultar saldo:", os.Args[0], "<maquina> saldo <id_conta>")
			return
		}
		idConta := parseInt(os.Args[3])
		var saldo float64
		err = client.Call("ServicoContas.ConsultarSaldo", idConta, &saldo)
		if err != nil {
			fmt.Printf("Erro ao consultar saldo da conta ID %d: %v\n", idConta, err)
		} else {
			fmt.Printf("Saldo da conta ID %d: %.2f\n", idConta, saldo)
		}

	default:
		fmt.Println("Operação inválida. Use: abrir, fechar, depositar, sacar, saldo")
	}
}

// Função auxiliar para converter string para float64
func parseFloat(arg string) float64 {
	val, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		fmt.Println("Valor inválido:", arg)
		os.Exit(1)
	}
	return val
}

// Função auxiliar para converter string para int
func parseInt(arg string) int {
	val, err := strconv.Atoi(arg)
	if err != nil {
		fmt.Println("Valor inválido:", arg)
		os.Exit(1)
	}
	return val
}
