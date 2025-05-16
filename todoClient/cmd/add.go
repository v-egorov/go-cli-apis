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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:          "add <Task>",
	Short:        "Добавить дело в список",
	SilenceUsage: true,
	Args:         cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiRoot := viper.GetString("api-root")
		return addAction(os.Stdout, apiRoot, args)
	},
}

func init() {
	log.Println("addCmd: init - AddCommand")
	rootCmd.AddCommand(addCmd)
}

func addAction(out io.Writer, apiRoot string, args []string) error {
	log.Println("addAction")
	task := strings.Join(args, " ")

	if err := addItem(apiRoot, task); err != nil {
		return err
	}

	return printAdd(out, task)
}

func printAdd(out io.Writer, task string) error {
	log.Println("printAdd")
	_, err := fmt.Fprintf(out, "Добавляем дело %q в список\n", task)
	return err
}
