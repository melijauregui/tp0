package common

import (
	"encoding/csv"
	"os"
	"strconv"
	"time"
)

// STORAGE_FILEPATH is the file where bets are stored.
const storageFilePath = "./bets.csv"

// LOTTERY_WINNER_NUMBER is the simulated winner number in the lottery contest.
const lotteryWinnerNumber = 7574

// Bet represents a lottery bet registry.
// agency must be passed as a string that can be converted to int,
// birthdate must be in the format "YYYY-MM-DD",
// number must be passed as a string that can be converted to int.
type Bet struct {
	Agency    int
	FirstName string
	LastName  string
	Document  string
	Birthdate time.Time
	Number    int
}

// NewBet creates a new Bet from string parameters.
// It converts agency and number to int and parses the birthdate.
func NewBet(agency, firstName, lastName, document, birthdate, number string) (Bet, error) {
	a, err := strconv.Atoi(agency)
	if err != nil {
		return Bet{}, err
	}

	num, err := strconv.Atoi(number)
	if err != nil {
		log.Errorf("action: failed to parse number: %v | result: fail ", err)
		return Bet{}, err
	}

	bd, err := time.Parse("2006-01-02", birthdate)
	if err != nil {
		log.Errorf("action: failed to parse birthdate: %v | result: fail ", err)
		return Bet{}, err
	}

	return Bet{
		Agency:    a,
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: bd,
		Number:    num,
	}, nil
}

// HasWon checks whether a bet won the prize or not.
func HasWon(bet Bet) bool {
	return bet.Number == lotteryWinnerNumber
}

// StoreBets persists the information of each bet in the storage file.
// Not thread-safe/process-safe.
func StoreBets(bets []Bet) error {
	file, err_opening := os.OpenFile(storageFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err_opening != nil {
		return err_opening
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	for _, bet := range bets {
		row := []string{
			strconv.Itoa(bet.Agency),
			bet.FirstName,
			bet.LastName,
			bet.Document,
			bet.Birthdate.Format("2006-01-02"),
			strconv.Itoa(bet.Number),
		}
		err_writing := writer.Write(row)
		if err_writing != nil {
			return err_writing
		}
	}
	writer.Flush()
	err_closing := writer.Error()
	if err_closing != nil {
		return err_closing
	}
	return nil
}

// LoadBets loads all the bets from the storage file.
// Not thread-safe/process-safe.
func LoadBets() ([]Bet, error) {
	file, err_opening := os.Open(storageFilePath)
	if err_opening != nil {
		return nil, err_opening
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err_reading := reader.ReadAll()
	if err_reading != nil {
		return nil, err_reading
	}

	var bets []Bet
	for _, row := range records {
		if len(row) < 6 {
			// Skip rows that do not have all required fields.
			continue
		}
		bet, err_creating_bet := NewBet(row[0], row[1], row[2], row[3], row[4], row[5])
		if err_creating_bet != nil {
			return nil, err_creating_bet
		}
		bets = append(bets, bet)
	}
	return bets, nil
}
