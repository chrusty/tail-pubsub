package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/cihub/seelog"

	context "golang.org/x/net/context"
	google "golang.org/x/oauth2/google"
	pubsub "google.golang.org/api/pubsub/v1"
)

var (
	project   = flag.String("project", "", "Google project-name")
	topic     = flag.String("topic", "", "PUB/SUB topic name to subscribe to")
	batchsize = flag.Int64("batchsize", 1, "How many messages to get at once")
)

func init() {
	// Parse the command-line arguments:
	flag.Parse()
}

func main() {
	// Make sure we flush the log before quitting:
	defer log.Flush()

	// Connect to GCE (either from GCE permissions, JSON file, or ENV-vars):
	client, err := google.DefaultClient(context.Background(), pubsub.PubsubScope)
	if err != nil {
		log.Criticalf("Unable to authenticate to GCE! (%s)", err)
		os.Exit(2)
	}

	// Get a DNS service-object:
	pubsubService, err := pubsub.New(client)
	if err != nil {
		log.Criticalf("Failed to connect to PUB/SUB! %v", err)
		os.Exit(2)
	}

	// Build the topic and subscription names:
	fullTopicName := fmt.Sprintf("projects/%v/topics/%v", *project, *topic)
	fullSubscriptionName := fmt.Sprintf("projects/%v/subscriptions/%v", *project, "tail-pubsub")

	// Create a subscription:
	log.Debugf("Making subscription to topic (%v) ...", fullTopicName)
	sub := &pubsub.Subscription{Topic: fullTopicName}
	subscription, err := pubsubService.Projects.Subscriptions.Create(fullSubscriptionName, sub).Do()
	if err != nil {
		log.Warnf("Failed to create subscription to topic (%v): %v", fullTopicName, err)
		log.Debugf("Getting existing subscription (%v) to topic (%v) ...", fullSubscriptionName, fullTopicName)
		subscription, err = pubsubService.Projects.Subscriptions.Get(fullSubscriptionName).Do()
		if err != nil {
			log.Criticalf("Couldn't even get existing subscription (%v): %v", fullSubscriptionName, err)
			os.Exit(2)
		}
	}
	log.Debugf("Subscription (%s) was created", subscription.Name)

	// Prepare a pull-request:
	pullRequest := &pubsub.PullRequest{
		ReturnImmediately: true,
		MaxMessages:       *batchsize,
	}

	// Now attack the pub/sub API:
	for {
		log.Debugf("Polling for messages ...")
		pullResponse, err := pubsubService.Projects.Subscriptions.Pull(fullSubscriptionName, pullRequest).Do()
		if err != nil {
			log.Errorf("Failed to pull messages from subscription (%v): %v", fullSubscriptionName, err)
		} else {
			for _, receivedMessage := range pullResponse.ReceivedMessages {
				messageData, err := base64.StdEncoding.DecodeString(receivedMessage.Message.Data)
				if err != nil {
					log.Warnf("Failed to decode message: %v", err)
				} else {
					log.Infof("Message: %v", messageData)
					ackRequest := &pubsub.AcknowledgeRequest{
						AckIds: []string{receivedMessage.AckId},
					}
					if _, err = pubsubService.Projects.Subscriptions.Acknowledge(fullSubscriptionName, ackRequest).Do(); err != nil {
						log.Warnf("Failed to acknowledge messages: %v", err)
					}
				}
			}
		}

		// Sleep until trying again:
		time.Sleep(1 * time.Second)

	}

}
