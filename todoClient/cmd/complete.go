/*
Copyright © 2025 Vladimir Egorov

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// completeCmd represents the complete command
var completeCmd = &cobra.Command{
	Use:           "complete",
	Short:         "Отметить дело <n> как завершенное",
	Args:          cobra.ExactValidArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("completeCmd RunE")
		apiRoot := viper.GetString("api-root")
		return completeAction(os.Stdout, apiRoot, args[0])
	},
}

func init() {
	log.Println("completeCmd: init - AddCommand")
	rootCmd.AddCommand(completeCmd)
}

func completeAction(out io.Writer, apiRoot, arg string) error {
	log.Printf("completeAction: aprRoot: %s, arg: %s", apiRoot, arg)

	id, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("%w: не является целым числом", ErrNotNumber)
	}

	if err := completeItem(apiRoot, id); err != nil {
		return err
	}
	return printComplete(out, id)
}

func printComplete(out io.Writer, id int) error {
	log.Printf("Item %d marked as completed\n", id)
	_, err := fmt.Fprintf(out, "Дело %d отмечено как выполненное\n", id)
	return err
}
