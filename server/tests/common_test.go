package main

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/common"
)

const storageFilePath = "./bets.csv"
const lotteryWinnerNumber = 7574

// TestBetInitMustKeepFields tests that a Bet is initialized correctly.
func TestBetInitMustKeepFields(t *testing.T) {
	// Cleanup after test.
	t.Cleanup(func() {
		os.Remove(storageFilePath)
	})

	bet, err := common.NewBet("1", "first", "last", "10000000", "2000-12-20", "7500")
	if err != nil {
		t.Fatalf("Error creating Bet: %v", err)
	}

	if bet.Agency != 1 {
		t.Errorf("Expected agency 1, got %d", bet.Agency)
	}
	if bet.FirstName != "first" {
		t.Errorf("Expected first name 'first', got %s", bet.FirstName)
	}
	if bet.LastName != "last" {
		t.Errorf("Expected last name 'last', got %s", bet.LastName)
	}
	if bet.Document != "10000000" {
		t.Errorf("Expected document '10000000', got %s", bet.Document)
	}

	expectedDate, _ := time.Parse("2006-01-02", "2000-12-20")
	if !bet.Birthdate.Equal(expectedDate) {
		t.Errorf("Expected birthdate %v, got %v", expectedDate, bet.Birthdate)
	}
	if bet.Number != 7500 {
		t.Errorf("Expected number 7500, got %d", bet.Number)
	}
}

// TestHasWonWinnerTrue tests that a bet with the winning number returns true.
func TestHasWonWinnerTrue(t *testing.T) {
	t.Cleanup(func() {
		os.Remove(storageFilePath)
	})

	bet, err := common.NewBet("1", "first", "last", "10000000", "2000-12-20", strconv.Itoa(lotteryWinnerNumber))
	if err != nil {
		t.Fatalf("Error creating Bet: %v", err)
	}
	if !common.HasWon(bet) {
		t.Errorf("Expected HasWon to return true for winning bet")
	}
}

// TestStoreBetsAndLoadBetsKeepsFieldsData tests that storing and loading bets preserves field data.
func TestStoreBetsAndLoadBetsKeepsFieldsData(t *testing.T) {
	t.Cleanup(func() {
		os.Remove(storageFilePath)
	})

	bet1, err := common.NewBet("1", "first", "last", "10000000", "2000-12-20", "7500")
	if err != nil {
		t.Fatalf("Error creating Bet: %v", err)
	}
	betsToStore := []common.Bet{bet1}

	if err := common.StoreBets(betsToStore); err != nil {
		t.Fatalf("Error storing bets: %v", err)
	}

	loadedBets, err := common.LoadBets()
	if err != nil {
		t.Fatalf("Error loading bets: %v", err)
	}
	if len(loadedBets) != 1 {
		t.Errorf("Expected 1 bet, got %d", len(loadedBets))
	}
	assertEqualBet(t, betsToStore[0], loadedBets[0])
}

// TestStoreBetsAndLoadBetsKeepsRegistryOrder tests that the order of bets is preserved.
func TestStoreBetsAndLoadBetsKeepsRegistryOrder(t *testing.T) {
	t.Cleanup(func() {
		os.Remove(storageFilePath)
	})

	bet0, err := common.NewBet("0", "first_0", "last_0", "10000000", "2000-12-20", "7500")
	if err != nil {
		t.Fatalf("Error creating Bet: %v", err)
	}
	bet1, err := common.NewBet("1", "first_1", "last_1", "10000001", "2000-12-21", "7501")
	if err != nil {
		t.Fatalf("Error creating Bet: %v", err)
	}
	betsToStore := []common.Bet{bet0, bet1}

	if err := common.StoreBets(betsToStore); err != nil {
		t.Fatalf("Error storing bets: %v", err)
	}

	loadedBets, err := common.LoadBets()
	if err != nil {
		t.Fatalf("Error loading bets: %v", err)
	}
	if len(loadedBets) != 2 {
		t.Errorf("Expected 2 bets, got %d", len(loadedBets))
	}
	assertEqualBet(t, betsToStore[0], loadedBets[0])
	assertEqualBet(t, betsToStore[1], loadedBets[1])
}

// assertEqualBet compares two Bet structs field by field.
func assertEqualBet(t *testing.T, b1, b2 common.Bet) {
	if b1.Agency != b2.Agency {
		t.Errorf("Agency mismatch: expected %d, got %d", b1.Agency, b2.Agency)
	}
	if b1.FirstName != b2.FirstName {
		t.Errorf("FirstName mismatch: expected %s, got %s", b1.FirstName, b2.FirstName)
	}
	if b1.LastName != b2.LastName {
		t.Errorf("LastName mismatch: expected %s, got %s", b1.LastName, b2.LastName)
	}
	if b1.Document != b2.Document {
		t.Errorf("Document mismatch: expected %s, got %s", b1.Document, b2.Document)
	}
	if !b1.Birthdate.Equal(b2.Birthdate) {
		t.Errorf("Birthdate mismatch: expected %v, got %v", b1.Birthdate, b2.Birthdate)
	}
	if b1.Number != b2.Number {
		t.Errorf("Number mismatch: expected %d, got %d", b1.Number, b2.Number)
	}
}
