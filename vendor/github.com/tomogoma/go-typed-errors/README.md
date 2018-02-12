# Go-Typed-Errors

Go-Typed-Errs is a go package that provides a mechanism for passing
on the class of the error being returned.

Common error classes are Auth(Forbidden/Unauthorized), Client and NotFound.

## Installation and Docs

Install using `go get github.com/tomogoma/go-typed-errors`

Godocs can be found at http://godoc.org/github.com/tomogoma/go-typed-errors

## Typical Usage

```golang


	// embed relevant 'Checkers' in struct
	ms := struct {
		typederrs.NotFoundErrCheck
		typederrs.NotImplErrCheck

		// do something returns an error which can be checked for type
		doSomething func() error
	}{
		doSomething: func() error {
			// return a typed error
			return typederrs.NewNotFoundf("something went wrong %s", "here")
		},
	}

	err := ms.doSomething()
	if err != nil {

		if ms.IsNotFoundError(err) {
			fmt.Println("resource not found")
			return
		}

		if ms.IsNotImplementedError(err) {
			fmt.Println("logic not implemented")
			return
		}

		// Act on generic error
		fmt.Printf("got a generic error: %v\n", err)
	}

	// Output: resource not found
```