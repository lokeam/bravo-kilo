export const queryKeys = {
  library: {
    root: ['library'] as const,
    page: {
      all: () => [...queryKeys.library.root, 'page'] as const,
      byDomain: (domain: string) => [...queryKeys.library.page.all(), domain] as const,
    }
  },
  home: {
    root: ['home'] as const,
    page: {
      all: () => [...queryKeys.home.root, 'page'] as const,
      byDomain: (domain: string) => [...queryKeys.home.page.all(), domain] as const,
    }
  },
  books: {
    root: ['books'] as const,
    all: (userID: number) => [...queryKeys.books.root, userID] as const,
    byId: (bookID: string) => [...queryKeys.books.root, 'detail', bookID] as const,
    byTitle: (title: string) => [...queryKeys.books.root, 'title', title] as const,
    authors: (userID: number) => [...queryKeys.books.root, 'authors', userID] as const,
    formats: (userID: number) => [...queryKeys.books.root, 'formats', userID] as const,
    genres: (userID: number) => [...queryKeys.books.root, 'genres', userID] as const,
    tags: (userID: number) => [...queryKeys.books.root, 'tags', userID] as const,

  },
} as const;