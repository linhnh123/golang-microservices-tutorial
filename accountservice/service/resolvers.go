package service

import "github.com/graphql-go/graphql"

type GraphQLResolvers interface {
	AccountResolverFunc(p graphql.ResolveParams) (interface{}, error)
}

type LiveGraphQLResolvers struct {
}

func (gqlres *LiveGraphQLResolvers) AccountResolverFunc(p graphql.ResolveParams) (interface{}, error) {
	account, err := fetchAccount(p.Context, p.Args["id"].(string))
	if err != nil {
		return nil, err
	}
	return account, nil
}
