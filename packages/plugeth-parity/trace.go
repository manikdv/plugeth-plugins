package main

import (
	"context"
	"strings"

	"github.com/openrelayxyz/plugeth-utils/core"
)

type InnerResult struct {
	Address string `json:"address,omitempty"`
	Code    string `json:"code,omitempty"`
	GasUsed string `json:"gasUsed,omitempty"`
	Output  string `json:"output,omitempty"`
}

type Action struct {
	CallType      string `json:"callType,omitempty"`
	From          string `json:"from,omitempty"`
	Address       string `json:"address,omitempty"`
	Balance       string `json:"balance,omitempty"`
	Gas           string `json:"gas,omitempty"`
	Init          string `json:"init,omitempty"`
	Input         string `json:"input,omitempty"`
	To            string `json:"to,omitempty"`
	RefundAddress string `json:"refundAddress,omitempty"`
	Value         string `json:"value,omitempty"`
}

type ParityResult struct {
	Action        *Action      `json:"action"`
	Error         string       `json:"error,omitempty"`
	Result        *InnerResult `json:"result,omitempty"`
	SubTraces     int          `json:"subtraces"`
	TracerAddress []int        `json:"traceAddress"`
	Type          string       `json:"type"`
}

type GethResponse struct {
	Type    string         `json:"type,omitempty"`
	From    string         `json:"from,omitempty"`
	To      string         `json:"to,omitempty"`
	Value   string         `json:"value,omitempty"`
	Gas     string         `json:"gas,omitempty"`
	GasUsed string         `json:"gasUsed,omitempty"`
	Input   string         `json:"input,omitempty"`
	Output  string         `json:"output,omitempty"`
	Error   string         `json:"error,omitempty"`
	Calls   []GethResponse `json:"calls,omitempty"`
}

type OuterGethResponse struct {
	Result GethResponse `json:"result"`
}

func FilterPrecompileCalls(calls []GethResponse) []GethResponse {
	result := []GethResponse{}
	for _, call := range calls {
		//develop test case to see what parity does if is legit call to procompiled contract
		if !strings.HasPrefix(call.To, "0x000000000000000000000000000000000000") || call.Value != "" {
			result = append(result, call)
		}
	}
	return result
}

func GethParity(gr GethResponse, address []int, t string) []*ParityResult {
	result := []*ParityResult{}
	calls := FilterPrecompileCalls(gr.Calls)
	addr := make([]int, len(address))
	copy(addr[:], address)

	if string(gr.GasUsed) == "" {
		gr.GasUsed = "0x0"
	}
	if gr.Output == "" {
		gr.Output = "0x"
	}
	if gr.Value == "" {
		gr.Value = "0x0"
	}

	unique := 0
	if gr.Error == "execution reverted" {
		unique = 1
	} else if gr.Error == "contract creation code storage out of gas" {
		unique = 3
	} else if gr.Error == "max code size exceeded" {
		unique = 3
	} else if gr.Error == "out of gas" {
		unique = 3
	} else if strings.HasPrefix(gr.Error, "gas") {
		unique = 3
	} else if strings.HasPrefix(gr.Error, "stack underflow") {
		unique = 5
	} else if strings.HasPrefix(gr.Error, "invalid opcode") {
		unique = 6
  } else if gr.Error == "invalid jump destination" {
		unique = 7
	} else if gr.Type == "CREATE" || gr.Type == "CREATE2" {
		unique = 2
	} else if gr.Type == "SELFDESTRUCT" {
		unique = 4
	} else if gr.Type == "STATICCALL" || gr.Type == "CALL" || gr.Type == "DELEGATECALL" {
		unique = 8
	}

	switch unique {
	case 0:
		result = append(result, &ParityResult{
			Action: &Action{
				CallType: strings.ToLower(gr.Type),
				From:     gr.From,
				Gas:      gr.Gas,
				Input:    gr.Input,
				To:       gr.To,
				Value:    gr.Value},
			Result: &InnerResult{GasUsed: gr.GasUsed,
				Output: gr.Output},
			SubTraces:     len(calls),
			TracerAddress: addr,
			Type:          t})

	case 1:
		result = append(result, &ParityResult{
			Action: &Action{
				CallType: strings.ToLower(gr.Type),
				From:     gr.From,
				Gas:      gr.Gas,
				Input:    gr.Input,
				Init:     gr.Input,
				To:       gr.To,
				Value:    gr.Value},
			Error:         "Reverted",
			SubTraces:     len(calls),
			TracerAddress: addr,
			Type:          strings.ToLower(gr.Type)})

	case 2:
		result = append(result, &ParityResult{
			Action: &Action{
				From:  gr.From,
				Gas:   gr.Gas,
				Init:  gr.Input,
				Value: gr.Value},
			Result: &InnerResult{
				Address: gr.To,
				Code:    gr.Output,
				GasUsed: gr.GasUsed,
			},
			SubTraces:     len(calls),
			TracerAddress: addr,
			Type:          "create"})

	case 3:
		result = append(result, &ParityResult{
			Action: &Action{
				CallType: strings.ToLower(gr.Type),
				From:  gr.From,
				To:    gr.To,
				Gas:   gr.Gas,
				Init:  gr.Input,
				Input: gr.Input,
				Value: gr.Value},
			Error:         "Out of gas",
			SubTraces:     len(calls),
			TracerAddress: addr,
			Type:          t})

	case 4:
		balance := gr.Value
		result = append(result, &ParityResult{
			Action: &Action{
				Address:       gr.From,
				Balance:       balance,
				RefundAddress: gr.To},
			Result:        &InnerResult{},
			SubTraces:     len(calls),
			TracerAddress: addr,
			Type:          "suicide"})

	case 5:
		result = append(result, &ParityResult{
			Action: &Action{
			CallType: strings.ToLower(gr.Type),
				From:     gr.From,
				Gas:      gr.Gas,
				Input:    gr.Input,
				To:       gr.To,
				Value:    gr.Value},
			Error:         "Stack undeflow",
			SubTraces:     len(calls),
			TracerAddress: addr,
			Type:          t})

	case 6:
		result = append(result, &ParityResult{
			Action: &Action{
				CallType: strings.ToLower(gr.Type),
				From:     gr.From,
				Gas:      gr.Gas,
				Input:    gr.Input,
				To:       gr.To,
				Value:    gr.Value},
			Error:         "Bad instruction",
			SubTraces:     len(calls),
			TracerAddress: addr,
			Type:          t})

	case 7:
		result = append(result, &ParityResult{
			Action: &Action{
				CallType: strings.ToLower(gr.Type),
				From:     gr.From,
				Gas:      gr.Gas,
				Input:    gr.Input,
				To:       gr.To,
				Value:    gr.Value},
			Error:         "Bad jump destination",
			SubTraces:     len(calls),
			TracerAddress: addr,
			Type:          t})

	case 8:
		result = append(result, &ParityResult{
			Action: &Action{
				CallType: strings.ToLower(gr.Type),
				From:     gr.From,
				Gas:      gr.Gas,
				Input:    gr.Input,
				To:       gr.To,
				Value:    gr.Value},
			Result: &InnerResult{
				Output:  gr.Output,
				GasUsed: gr.GasUsed,
			},
			SubTraces:     len(calls),
			TracerAddress: addr,
			Type:          "call"})
	}

	for i, call := range calls {
		if call.Type == "DELEGATECALL" {
			call.Value = gr.Value
		}
		result = append(result, GethParity(call, append(address, i), t)...)
	}
	return result
}

func (tr *ParityTrace) TraceVariantCall(ctx context.Context, txObject map[string]interface{}, bkNum string) ([]*ParityResult, string, error) {
	client, err := tr.stack.Attach()
	if err != nil {
		return nil, "", err
	}
	gr := GethResponse{}
	client.Call(&gr, "debug_traceCall", txObject, bkNum, map[string]string{"tracer": "callTracer"})
	tAddress := make([]int, 0)
	gp := GethParity(gr, tAddress, strings.ToLower(gr.Type))
	if gr.Output == "" {
		gr.Output = "0x"
	}
	output := gr.Output
	trace := gp
	return trace, output, err
}

func (tr *ParityTrace) TraceVariantTransaction(ctx context.Context, txHash core.Hash) ([]*ParityResult, string, error) {
	client, err := tr.stack.Attach()
	if err != nil {
		return nil, "", err
	}
	gr := GethResponse{}
	client.Call(&gr, "debug_traceTransaction", txHash, map[string]string{"tracer": "callTracer"})
	tAddress := make([]int, 0)
	gp := GethParity(gr, tAddress, strings.ToLower(gr.Type))
	if gr.Output == "" {
		gr.Output = "0x"
	}
	output := gr.Output
	trace := gp
	return trace, output, err
}

func (tr *ParityTrace) TraceVariantBlock(ctx context.Context, bkNum string) ([][]*ParityResult, []string, error) {
	client, err := tr.stack.Attach()
	if err != nil {
		return nil, nil, err
	}
	outputs := []string{}
	gr := []OuterGethResponse{}
	err = client.Call(&gr, "debug_traceBlockByNumber", bkNum, map[string]string{"tracer": "callTracer"})
	if err != nil {
		return nil, nil, err
	}
	pr := [][]*ParityResult{}
	for _, item := range gr {
		outputs = append(outputs, item.Result.Output)
		tAddress := make([]int, 0)
		pr = append(pr, GethParity(item.Result, tAddress, strings.ToLower(item.Result.Type)))
	}

	result := pr
	return result, outputs, err
}
