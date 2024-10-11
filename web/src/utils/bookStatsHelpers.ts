import { Book, HomepageStatistics } from './../types/api';

export const updateBookStats = (
  stats: Array<{ label: string; count: number}>,
  newItems: string[],
): Array<{ label: string; count: number }> => {
  const statsMap = new Map(stats.map(item => [item.label, item]));

  for (let item of newItems) {
    const currentValue = statsMap.get(item);
    statsMap.set(item, { label: item, count: (currentValue?.count || 0) + 1});
  }

  return Array.from(statsMap.values());
};

export const updateHomepageStats = (
  oldStats: HomepageStatistics,
  newBook: Book
): HomepageStatistics => {
  return {
    ...oldStats,
    userBkGenres: {
      booksByGenre: updateBookStats(oldStats.userBkGenres.booksByGenre, newBook.genres),
    },
    userAuthors: {
      booksByAuthor: updateBookStats(oldStats.userAuthors.booksByAuthor, newBook.authors),
    },
    userBkLang: {
      booksByLang: updateBookStats(oldStats.userBkLang.booksByLang, [newBook.language]),
    },
    userTags: {
      userTags: updateBookStats(oldStats.userTags.userTags, newBook.tags || []),
    },
  };
}
