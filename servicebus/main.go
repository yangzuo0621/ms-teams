package main

import (
	"context"
	"fmt"

	"github.com/Azure/azure-amqp-common-go/v3/aad"
	servicebus "github.com/Azure/azure-service-bus-go"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

const (
	tenantID     = ""
	clientID     = ""
	clientSecret = ""
)

func main() {

	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, tenantID)
	if err != nil {
		panic(err)
	}

	token, err := adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, azure.PublicCloud.ResourceIdentifiers.ServiceBus)
	if err != nil {
		panic(err)
	}

	// if err := token.Refresh(); err != nil {
	// 	panic(err)
	// }

	// ns, err := servicebus.NewNamespace(servicebus.NamespaceWithEnvironmentBinding("e2eaksbus-zuyaebld38219185"))

	// tokenProvider := aad.JWTProviderWithAADToken(token)
	tokenProvider, err := aad.NewJWTProvider(
		aad.JWTProviderWithAADToken(token),
	)
	if err != nil {
		panic(err)
	}

	ns, err := servicebus.NewNamespace(
		servicebus.NamespaceWithAzureEnvironment("e2eaksbus-zuyaebld38219185", "AZUREPUBLICCLOUD"),
		servicebus.NamespaceWithTokenProvider(tokenProvider),
	)
	if err != nil {
		panic(err)
	}

	// ns.Name = "e2eaksbus-zuyaebld38219185"

	// ns, err := servicebus.NewNamespace(servicebus.NamespaceWithEnvironmentBinding("e2eaksbus-zuyaebld38219185"))
	// if err != nil {
	// 	panic(err)
	// }

	ctx := context.Background()
	qm := ns.NewQueueManager()
	qe, err := ensureQueue(ctx, qm, "MessageBatchingExample")
	if err != nil {
		fmt.Println(err)
		return
	}

	q, err := ns.NewQueue(qe.Name)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		_ = q.Close(ctx)
	}()

	msgs := make([]*servicebus.Message, 10)
	for i := 0; i < 10; i++ {
		msgs[i] = servicebus.NewMessageFromString(fmt.Sprintf("foo %d", i))
	}

	batcher := servicebus.NewMessageBatchIterator(servicebus.StandardMaxMessageSizeInBytes, msgs...)
	if err := q.SendBatch(ctx, batcher); err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < 10; i++ {
		err := q.ReceiveOne(ctx, MessagePrinter{})
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// authorizer := autorest.NewBearerAuthorizer(token)

	// if err != nil {
	// 	panic(err)
	// }

	// spClient := graphrbac.NewServicePrincipalsClient(tenantID)
	// spClient.Authorizer = authorizer

	// // ""
	// results, err := spClient.List(context.Background(), fmt.Sprintf("appId eq '%s'", "20263b94-2231-48e8-8948-88dbf6f24c80"))
	// if err != nil {
	// 	panic(err)
	// }

	// for _, v := range results.Values() {
	// 	fmt.Println(*v.AppID, *v.ObjectID, *v.DisplayName)
	// }

}

type MessagePrinter struct{}

func (mp MessagePrinter) Handle(ctx context.Context, msg *servicebus.Message) error {
	fmt.Println(string(msg.Data))
	return msg.Complete(ctx)
}

func ensureQueue(ctx context.Context, qm *servicebus.QueueManager, name string, opts ...servicebus.QueueManagementOption) (*servicebus.QueueEntity, error) {
	qe, err := qm.Get(ctx, name)
	if err == nil {
		_ = qm.Delete(ctx, name)
	}

	qe, err = qm.Put(ctx, name, opts...)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return qe, nil
}
