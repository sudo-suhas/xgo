package httputil_test

import (
	"fmt"

	"github.com/sudo-suhas/xgo/httputil"
)

func ExampleURLBuilderSource() {
	b, err := httputil.NewURLBuilderSource("https://api.example.com/v1")
	check(err)

	u := b.NewURLBuilder().
		Path("/users").
		URL()
	fmt.Println("URL:", u.String())

	// Output:
	// URL: https://api.example.com/v1/users
}

func ExampleURLBuilder() {
	b, err := httputil.NewURLBuilderSource("https://api.example.com/v1/")
	check(err)

	u := b.NewURLBuilder().
		Path("/users/{userID}/posts/{postID}/comments").
		PathParam("userID", "foo").
		PathParamInt("postID", 42).
		QueryParam("author_id", "foo", "bar").
		QueryParamInt("limit", 20).
		QueryParamBool("recent", true).
		URL()
	fmt.Println("URL:", u.String())

	u = b.NewURLBuilder().
		Path("posts/{title}").
		PathParam("title", `Letters & "Special" Characters`).
		URL()
	fmt.Println("URL (encoded path param):", u.String())

	// Output:
	// URL: https://api.example.com/v1/users/foo/posts/42/comments?author_id=foo&author_id=bar&limit=20&recent=true
	// URL (encoded path param): https://api.example.com/v1/posts/Letters%2520&%2520%2522Special%2522%2520Characters
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
