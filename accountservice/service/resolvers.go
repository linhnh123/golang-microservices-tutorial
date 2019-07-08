package service

import (
	"github.com/graphql-go/graphql"
	internalmodel "github.com/linhnh123/golang-microservices-tutorial/accountservice/model"
)

type GraphQLResolvers interface {
	AccountResolverFunc(p graphql.ResolveParams) (interface{}, error)
}

type LiveGraphQLResolvers struct {
}

func (gqlres *LiveGraphQLResolvers) AccountResolverFunc(p graphql.ResolveParams) (interface{}, error) {
	// account, err := fetchAccount(p.Context, p.Args["id"].(string))
	// if err != nil {
	// 	return nil, err
	// }
	account := internalmodel.Account{
		Id:       "1234",
		Name:     "Test name",
		ServedBy: "Served",
	}
	return account, nil
}
