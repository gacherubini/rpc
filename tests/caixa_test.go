package main

import (
	"fmt"
	"net/rpc"
	"sync"
	"testing"

	"github.com/google/uuid"
)

type DepositoArgs struct {
	IDConta     int
	Valor       float64
	TransacaoID string
}

type SacarArgs struct {
	IDConta     int
	Valor       float64
	TransacaoID string
}

func gerarTransacaoID() string {
	return uuid.New().String()
}

func TestConcorrenciaDepositoSaque(t *testing.T) {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		t.Fatalf("Erro ao conectar ao servidor: %v", err)
	}
	defer client.Close()

	err = client.Call("ServicoContas.LimparDados", struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Erro ao limpar dados do servidor: %v", err)
	}

	var idConta int
	err = client.Call("ServicoContas.AbrirConta", 1000.0, &idConta)
	if err != nil {
		t.Fatalf("Erro ao abrir conta: %v", err)
	}

	var wg sync.WaitGroup
	// []

	for i := 0; i < 10; i++ { // 10 * 50 = 500
		wg.Add(1) // [1, 1, 1,]
		go func() {
			defer wg.Done()
			var resultado bool
			args := DepositoArgs{
				IDConta:     idConta,
				Valor:       50.0,
				TransacaoID: gerarTransacaoID(),
			}
			err := client.Call("ServicoContas.Depositar", args, &resultado)
			if err != nil || !resultado {
				t.Errorf("Erro ao depositar na conta ID %d: %v", idConta, err)
			}
		}()
	}

	for i := 0; i < 10; i++ { // 30 * 10 = 300
		wg.Add(1)
		go func() {
			defer wg.Done()
			var resultado bool
			args := SacarArgs{
				IDConta:     idConta,
				Valor:       30.0,
				TransacaoID: gerarTransacaoID(),
			}
			err := client.Call("ServicoContas.Sacar", args, &resultado)
			if err != nil || !resultado {
				t.Errorf("Erro ao sacar da conta ID %d: %v", idConta, err)
			}
		}()
	}

	wg.Wait()

	var saldoFinal float64
	err = client.Call("ServicoContas.ConsultarSaldo", idConta, &saldoFinal)
	if err != nil {
		t.Fatalf("Erro ao consultar saldo final: %v", err)
	}

	saldoEsperado := 1200.0
	if saldoFinal != saldoEsperado {
		t.Errorf("Saldo incorreto. Esperado: %.2f, Obtido: %.2f", saldoEsperado, saldoFinal)
	} else {
		fmt.Printf("Teste de concorrência de depósito e saque concluído com sucesso! Saldo final: %.2f\n", saldoFinal)
	}

	err = client.Call("ServicoContas.LimparDados", struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Erro ao limpar dados do servidor: %v", err)
	}
}

func TestIdempotenciaTransacao(t *testing.T) {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		t.Fatalf("Erro ao conectar ao servidor: %v", err)
	}
	defer client.Close()

	err = client.Call("ServicoContas.LimparDados", struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Erro ao limpar dados do servidor: %v", err)
	}

	var idConta int
	err = client.Call("ServicoContas.AbrirConta", 1000.0, &idConta)
	if err != nil {
		t.Fatalf("Erro ao abrir conta: %v", err)
	}

	transacaoID := gerarTransacaoID()
	args := DepositoArgs{IDConta: idConta, Valor: 100.0, TransacaoID: transacaoID}

	var resultado bool
	err = client.Call("ServicoContas.Depositar", args, &resultado) // dar certo,  transacao id 1
	if err != nil || !resultado {
		t.Fatalf("Erro ao realizar primeiro depósito: %v", err)
	}

	err = client.Call("ServicoContas.Depositar", args, &resultado) // dar errado, transacao id 1
	if err != nil || !resultado {
		t.Fatalf("Erro ao realizar segundo depósito (esperado ser ignorado): %v", err)
	}

	var saldoFinal float64
	err = client.Call("ServicoContas.ConsultarSaldo", idConta, &saldoFinal) // procura no proprio banco o saldo atual
	if err != nil {
		t.Fatalf("Erro ao consultar saldo final: %v", err)
	}

	saldoEsperado := 1100.0
	if saldoFinal != saldoEsperado {
		t.Errorf("Saldo incorreto. Esperado: %.2f, Obtido: %.2f", saldoEsperado, saldoFinal)
	} else {
		fmt.Printf("Teste de idempotência concluído com sucesso! Saldo final: %.2f\n", saldoFinal)
	}

	err = client.Call("ServicoContas.LimparDados", struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Erro ao limpar dados do servidor: %v", err)
	}
}

func TestIdempotenciaComFalhaDeRede(t *testing.T) {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		t.Fatalf("Erro ao conectar ao servidor: %v", err)
	}
	defer client.Close()

	err = client.Call("ServicoContas.LimparDados", struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Erro ao limpar dados do servidor: %v", err)
	}

	var idConta int
	err = client.Call("ServicoContas.AbrirConta", 1000.0, &idConta)
	if err != nil {
		t.Fatalf("Erro ao abrir conta: %v", err)
	}

	transacaoID := gerarTransacaoID()

	args := DepositoArgs{IDConta: idConta, Valor: 100.0, TransacaoID: transacaoID}
	var resultado bool

	err = client.Call("ServicoContas.Depositar", args, &resultado) // fez falhar(nao foi pro bd)
	if err == nil || !resultado {
		t.Logf("Falha simulada: desconectando o cliente durante a primeira tentativa.")
	}

	client, err = rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		t.Fatalf("Erro ao reconectar ao servidor após falha simulada: %v", err)
	}
	defer client.Close()

	err = client.Call("ServicoContas.Depositar", args, &resultado)
	if err != nil || !resultado {
		t.Fatalf("Erro ao realizar depósito após falha: %v", err)
	}

	var saldoFinal float64
	err = client.Call("ServicoContas.ConsultarSaldo", idConta, &saldoFinal)
	if err != nil {
		t.Fatalf("Erro ao consultar saldo final: %v", err)
	}

	saldoEsperado := 1100.0 // + 100
	if saldoFinal != saldoEsperado {
		t.Errorf("Saldo incorreto. Esperado: %.2f, Obtido: %.2f", saldoEsperado, saldoFinal)
	} else {
		fmt.Printf("Teste de idempotência com falha de rede concluído com sucesso! Saldo final: %.2f\n", saldoFinal)
	}

	err = client.Call("ServicoContas.LimparDados", struct{}{}, &struct{}{})
	if err != nil {
		t.Fatalf("Erro ao limpar dados do servidor: %v", err)
	}
}
