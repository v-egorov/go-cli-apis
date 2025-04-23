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
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:          "view",
	Short:        "Показать детали одной записи",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("viewCmd RunE")
		apiRoot := viper.GetString("api-root")
		return viewAction(os.Stdout, apiRoot, args[0])
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}

func viewAction(out io.Writer, apiRoot, arg string) error {
	log.Printf("vewAction: apiRoot: %s, arg: %s", apiRoot, arg)
	id, err := strconv.Atoi(arg)
	if err != nil {
		log.Printf("Invalid arg, must be int: %s", arg)
		return fmt.Errorf("%w: неверный номер дела: %s", ErrNotNumber, arg)
	}

	i, err := getOne(apiRoot, id)
	if err != nil {
		log.Printf("Error from apiRoot: %s", err)
		return err
	}
	return printOne(out, i)
}

const timeFormat = "Jan/02 @15:04"

func printOne(out io.Writer, i item) error {
	w := tabwriter.NewWriter(out, 14, 2, 0, ' ', 0)
	fmt.Fprintf(out, "Дело:\t\t%s\n", i.Task)
	fmt.Fprintf(out, "Создано:\t%s\n", i.CreatedAt.Format(timeFormat))
	if i.Done {
		fmt.Fprintf(out, "Завершено:\t%s\n", "Да")
		fmt.Fprintf(out, "\t\t%s\n", i.CompletedAt.Format(timeFormat))
		return w.Flush()
	}

	fmt.Fprintf(out, "Завершено:\t%s\n", "Нет")
	return w.Flush()
}
