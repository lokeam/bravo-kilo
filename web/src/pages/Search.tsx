import CardList from "../components/CardList/CardList";
import useStore from "../store/useStore";

const Search = () => {
  const { searchResults } = useStore();

  console.log('search page - nav results: ', searchResults);

  return (
    <div className="bk_lib flex flex-col px-5 antialiased md:px-1 md:ml-24 h-screen pt-20">

      <div className="flex flex-row relative w-full max-w-7xl justify-between items-center text-left text-white border-b-2 border-solid border-zinc-700 pb-6 mb-2">

      </div>
      {/* {sortedBooks && sortedBooks.length > 0 && <CardList books={sortedBooks} />} */}
    </div>
  )
}

export default Search
