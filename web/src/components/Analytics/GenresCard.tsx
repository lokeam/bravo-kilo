
function GenresCard() {

  const customers = [
    {
      id: '0',
      bookCount: '22',
      genreName: 'Classics',
    },
    {
      id: '1',
      bookCount: '18',
      genreName: 'Adventure',
    },
    {
      id: '2',
      bookCount: '13',
      genreName: 'Science Fiction',
    },
    {
      id: '3',
      bookCount: '10',
      genreName: 'Witty',
    },
    {
      id: '4',
      bookCount: '8',
      genreName: 'Mystery',
    },
    {
      id: '5',
      bookCount: '7',
      genreName: 'Political',
    },
    {
      id: '6',
      bookCount: '6',
      genreName: 'Inspiring',
    },
    {
      id: '6',
      bookCount: '3',
      genreName: 'Funny',
    },
  ];

  return (
    <div className=" col-span-full mdTablet:col-span-4 bg-maastricht shadow-sm rounded-xl">
      <header className="px-5 py-4 border-b border-gray-100 dark:border-gray-700/60">
        <h2 className="text-left text-lg font-semibold text-gray-800 dark:text-gray-100">Books By Genre</h2>
      </header>
      <div className="p-3">
        <div className="overflow-x-hidden">
          {/* Table header */}
          <ul className="flex place-content-between text-xs font-semibold uppercase text-gray-400 bg-opacity-50">
            <li>Genre Name</li>
            <li>Book Count</li>
          </ul>
          {/* Table body */}
          <ul className="text-sm divide-y leading-6 dark:divide-gray-700/60 mt-3 mb-4">
            {
              customers.map(customer => {
                return (
                  <li key={customer.id} className="relative border-none mb-2 p-1">
                    <div className="horizontal_bar w-full top-0 left-0 bg-hepatica-lt/[0.2] rounded absolute p-4"></div>
                    <div className="relative h-full flex flex-row place-content-between items-center px-2">
                      <div className="z-10 text-white text-base text-left flex flex-row gap-2">
                        <div className="text-white">{customer.genreName}</div>
                      </div>
                      <div className="text-base text-right">{customer.bookCount}</div>
                    </div>
                  </li>
                )
              })
            }
          </ul>
        </div>
        </div>
    </div>
  );
}

export default GenresCard;