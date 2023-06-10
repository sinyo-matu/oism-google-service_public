package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type DynamoDbClient struct {
	inner *dynamodb.Client
}

func (i DynamoDbClient) PutToken(userExId string, token *oauth2.Token) error {
	t, err := json.Marshal(token)
	if err != nil {
		return err
	}
	item := map[string]types.AttributeValue{
		"ID":       &types.AttributeValueMemberS{Value: uuid.New().String()},
		"UserExId": &types.AttributeValueMemberS{Value: userExId},
		"Tokens":   &types.AttributeValueMemberS{Value: string(t)},
	}
	_, err = i.inner.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(DynamoDbTableName),
		Item:      item,
	})
	if err != nil {
		return err
	}
	return nil
}

func (i DynamoDbClient) FetchToken(userExId string) (*oauth2.Token, error) {
	key := map[string]types.AttributeValue{
		"UserExId": &types.AttributeValueMemberS{Value: userExId},
	}
	output, err := i.inner.GetItem(context.TODO(), &dynamodb.GetItemInput{Key: key, TableName: aws.String(DynamoDbTableName)})
	if err != nil {
		return nil, err
	}
	var record DynamoDbUserExIdTokens
	if len(output.Item) == 0 {
		return nil, fmt.Errorf("no such userExId: %v", userExId)
	}
	err = attributevalue.UnmarshalMap(output.Item, &record)
	if err != nil {
		return nil, err
	}
	token, err := parseGoogleToken(record.Tokens)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func parseGoogleToken(j string) (*oauth2.Token, error) {
	var t *oauth2.Token
	err := json.Unmarshal([]byte(j), &t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func IsNotSuchKeyError(err error) bool {
	return strings.HasPrefix(err.Error(), "no such userExId:")
}
