package auth_test

import "testing"

func TestNew(t *testing.T) {

	// TODO save login details for all auth types
	// 1. Successful register
	// 2. Invalid (unsuccessful register)
	// 3. Successful login
	// 4. Invalid unsuccessful login
	// 5. Successful token validation
	// 6. Invalid (unsuccessful) token validation

	// TODO limit auth attempts when unsuccessful
	// 1. n login/token attempts from an ip address for a user over duration t
	// 2. m>n login/token attempts for a user over duration t
	// 3. p>m login/token attempts from an ip address over duration t

	// TODO enforce api keys
	// 1 client: 1 key (a client is a microservice or front-end app)
	// Access auth services only if client/key combo is recogonized

	t.Fatalf("Not yet implemented")
}

func TestAuth_RegisterUser(t *testing.T) {
}
