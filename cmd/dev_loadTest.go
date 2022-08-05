package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	pylonsApp "github.com/Pylons-tech/pylons/app"
	pb "github.com/Pylons-tech/pylons/x/pylons/types"
	"github.com/aliirns/cosmos-transaction-go/transaction"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	// 	"github.com/Pylons-tech/pylons/x/pylons/types

	//"github.com/olekukonko/tablewriter"
	"go.uber.org/atomic"
)

const (
	_KEYNAME    = 0
	_ADDRESS    = 1
	_PRIVATEKEY = 2
)

var wg sync.WaitGroup

func DevLoadTest() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "load-test [users] [grpc] [chainID]",
		Short:   "Simulate a load test given the number of users",
		Long:    "Simulate a load test given the number of users, the provided number of user accounts will be created and used to perform transactions via GRPC",
		Example: "pylonsd load-test 10000 127.0.0.1:9090 pylons-testnet-1",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			Users, err := strconv.Atoi(args[0])
			grpcURL := args[1]
			chainID := args[2]

			if err != nil {
				return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid number of users ", args[0])
			}

			var failureCount atomic.Uint32

			Accounts, err := GenerateAccounts(Users)
			if err != nil {
				return sdkerrors.Wrapf(sdkerrors.ErrLogic, "unable to generate accounts")
			}

			t1 := time.Now()

			wg.Add(Users)

			for i := 0; i < Users; i++ {
				msg := pb.MsgExecuteRecipe{Creator: Accounts[i][_ADDRESS], CookbookId: "cb130", RecipeId: "LOUDGetCharactercb130", CoinInputsIndex: 0, ItemIds: []string{}, PaymentInfos: []pb.PaymentInfo{}}
				go threadedLoadTest(Accounts[i][_KEYNAME], Accounts[i][_ADDRESS], Accounts[i][_PRIVATEKEY], &msg, &failureCount, chainID, grpcURL)
			}

			wg.Wait()

			elapsed := time.Since(t1)

			TPS := float64(Users) / (elapsed.Seconds())
			fmt.Println("Summary of LoadTest")
			fmt.Printf(" %v Concurrent transactions were performed\n %v transactions failed \n time taken : %v\n TPS Achieved : %v \n", Users, failureCount.String(), elapsed, TPS)

			return nil
		},
	}

	return cmd
}

func readCsvFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, err
}

func threadedLoadTest(myKey string, myaddress string, myprivateKey string, m sdk.Msg, atomicCounter *atomic.Uint32, chainID string, grpcURL string) {
	defer wg.Done()
	config := pylonsApp.DefaultConfig()
	res, err := transaction.CosmosTx(myaddress, myprivateKey, grpcURL, m, chainID, config)
	if err != nil {
		atomicCounter.Add(1)
		return
	}
	if res.TxResponse.Code != 0 {
		atomicCounter.Add(1)
		return
	}
	return

}

func GenerateAccounts(n int) ([][]string, error) {
	accounts, err := readCsvFile("TestAccounts.csv")
	if err != nil {

		fmt.Println("TestAccounts.csv not found ... generating new one")
	}
	numAccounts := len(accounts)

	if numAccounts >= n {
		return accounts, nil
	}

	fmt.Printf("Generating some more accounts ...\n")
	bash := "generateAccountstoCSV.sh"
	cmd := exec.Command("sh", []string{bash, strconv.Itoa(numAccounts), strconv.Itoa(n)}...)
	b, err := cmd.CombinedOutput()
	fmt.Println(string(b))
	if err != nil {
		return nil, err
	}

	newAccounts, err := readCsvFile("TestAccounts.csv")
	if err != nil {
		return nil, err
	}

	return newAccounts, err

}
