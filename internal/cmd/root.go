package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"

	"github.com/vvatanabe/dynamomq"

	"github.com/aws/aws-sdk-go-v2/aws"
)

var rootCmd = &cobra.Command{
	Use:   "dynamomq",
	Short: "dynamomq is a tool for implementing message queueing with Amazon DynamoDB in Go",
	Long: `dynamomq is a tool for implementing message queueing with Amazon DynamoDB in Go.

Environment Variables:
  * AWS_REGION
  * AWS_PROFILE
  * AWS_ACCESS_KEY_ID
  * AWS_SECRET_ACCESS_KEY
  * AWS_SESSION_TOKEN
  refs: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
`,
	Version: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		defer fmt.Printf("... Interactive is ending\n\n\n")

		fmt.Println("===========================================================")
		fmt.Println(">> Welcome to DynamoMQ CLI! [INTERACTIVE MODE]")
		fmt.Println("===========================================================")
		fmt.Println("for help, enter one of the following: ? or h or help")
		fmt.Println("all commands in CLIs need to be typed in lowercase")
		fmt.Println("")

		ctx := context.Background()
		client, cfg, err := createDynamoMQClient[any](ctx, flgs)
		if err != nil {
			return fmt.Errorf("... %v\n", err)
		}

		fmt.Println("... AWS session is properly established!")

		fmt.Printf("AWSRegion: %s\n", cfg.Region)
		fmt.Printf("TableName: %s\n", flgs.TableName)
		fmt.Printf("EndpointURL: %s\n", flgs.EndpointURL)
		fmt.Println("")

		c := Interactive{
			TableName: flgs.TableName,
			Client:    client,
			Message:   nil,
		}

		// 1. Create a Scanner using the InputStream available.
		scanner := bufio.NewScanner(os.Stdin)

		for {
			// 2. Don't forget to prompt the user
			if c.Message != nil {
				fmt.Printf("\nID <%s> >> Enter command: ", c.Message.ID)
			} else {
				fmt.Print("\n>> Enter command: ")
			}

			// 3. Use the Scanner to read a line of text from the user.
			scanned := scanner.Scan()
			if !scanned {
				break
			}

			input := scanner.Text()
			if input == "" {
				continue
			}

			command, params := parseInput(input)
			switch command {
			case "":
				continue
			case "quit", "q":
				return nil
			default:
				// 4. Now, you can do anything with the input string that you need to.
				// Like, output it to the user.
				c.Run(context.Background(), command, params)
			}
		}
		return nil
	},
}

func createDynamoMQClient[T any](ctx context.Context, flags *Flags) (dynamomq.Client[T], aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, cfg, fmt.Errorf("failed to load aws config: %s", err)
	}
	client, err := dynamomq.NewFromConfig[T](cfg,
		dynamomq.WithTableName(flags.TableName),
		dynamomq.WithQueueingIndexName(flags.QueueingIndexName),
		dynamomq.WithAWSBaseEndpoint(flags.EndpointURL))
	if err != nil {
		return nil, cfg, fmt.Errorf("AWS session could not be established!: %v", err)
	}
	return client, cfg, nil
}

func parseInput(input string) (command string, params []string) {
	input = strings.TrimSpace(input)
	arr := strings.Fields(input)

	if len(arr) == 0 {
		return "", nil
	}

	command = strings.ToLower(arr[0])

	if len(arr) > 1 {
		params = make([]string, len(arr)-1)
		for i := 1; i < len(arr); i++ {
			params[i-1] = strings.TrimSpace(arr[i])
		}
	}
	return command, params
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&flgs.TableName, flagMap.TableName.Name, flagMap.TableName.Value, flagMap.TableName.Usage)
	rootCmd.Flags().StringVar(&flgs.QueueingIndexName, flagMap.QueueingIndexName.Name, flagMap.QueueingIndexName.Value, flagMap.QueueingIndexName.Usage)
	rootCmd.Flags().StringVar(&flgs.EndpointURL, flagMap.EndpointURL.Name, flagMap.EndpointURL.Value, flagMap.EndpointURL.Usage)
}
