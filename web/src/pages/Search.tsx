import CardList from "../components/CardList/CardList";
import useStore from "../store/useStore";

const Search = () => {
  const { searchResults } = useStore();

  console.log('search page - nav results: ', searchResults);

  return (
    <div className="bk_lib flex flex-col px-5 antialiased md:px-1 md:ml-24 h-screen pt-20">

      <div className="flex flex-row relative w-full max-w-7xl justify-between items-center text-left text-white border-b-2 border-solid border-zinc-700 pb-6 mb-2">

      </div>
      {/*

        A. Make request
          a. Check if cached request exists, if yes move to D. (Render)
          b. Render loading indicator
          c. Get response
        B. Reformat response data (create separate hook for this)
        C. Save in Global Storage
          a. Remove loading indicator
        D. Render in Page

        --- This all must happen before caching data in Global Storage
        - Take nav results array
        - Create temp array to reformatted response
        - Look at each element in response array
          - Extract the volumeInfo obj
          - Create a new object holding all book data
            - Transfer all data from book fields into new obj
            - Check if new object title || ISBN10 || ISBN13 exists in user's allBooks array
              - If yes, then set existsInLibrary custom flag in CardItemSearch data model to true
          - Add new obj to reformatted response array

      {sortedBooks && sortedBooks.length > 0 && <CardList books={sortedBooks} />} */}
    </div>
  )
}

export default Search
