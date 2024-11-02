package redis

const (
	PrefixBook = "book:"                          // for HandleGetAllUserBooks
	PrefixAuthToken = "auth:"
	PrefixUserDelete = "user:delete:"
	PrefixDeletionQueue = "deletion:queue"

	PrefixBookDetail = "book:detail:"            // for HandleGetBookByID
	PrefixBookAuthor = "book:author:"            // for HandleGetBooksByAuthors
	PrefixBookFormat = "book:format:"            // for HandleGetBooksByFormat
	PrefixBookGenre = "book:genre:"              // for HandleGetBooksByGenres
	PrefixBookTag = "book:tag:"                  // for HandleGetBooksByTags
	PrefixBookHomepage = "book:homepage"         // for HandleGetHomepageData
)