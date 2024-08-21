
function LanguagesCard() {

  const customers = [
    {
      id: '0',
      bookCount: '12',
      flag: '🇺🇸',
      language: 'English (US)',
    },
    {
      id: '1',
      bookCount: '10',
      flag: '🇩🇪',
      language: 'Deutsch',
    },
    {
      id: '2',
      bookCount: '2',
      flag: '🇯🇵',
      language: '日本語',
    },
    {
      id: '3',
      bookCount: '5',
      flag: '🇫🇷',
      language: 'Français',
    },
    {
      id: '4',
      bookCount: '5',
      flag: '🇰🇷',
      language: '한국어',
    },
  ];

  return (
    <div className=" col-span-full lgMobile:col-span-6 mdTablet:col-span-4 bg-maastricht shadow-sm rounded-xl">
      <header className="px-5 py-4 border-b border-gray-100 dark:border-gray-700/60">
        <h2 className="text-left text-lg font-semibold text-gray-800 dark:text-gray-100">Books By Language</h2>
      </header>
      <div className="p-3">
        <div className="overflow-x-hidden">
          {/* Table header */}
          <ul className="flex place-content-between text-xs font-semibold uppercase text-gray-400 bg-opacity-50">
            <li>Language</li>
            <li>Book Count</li>
          </ul>
          {/* Table body */}
          <ul className="text-sm divide-y leading-6 dark:divide-gray-700/60 mt-3 mb-4">
            {
              customers.map(customer => {
                return (
                  <li key={customer.id} className="relative border-none mb-2 p-1">
                    <div className="horizontal_bar w-full top-0 left-0 bg-maya-blue/[0.2] rounded absolute p-4"></div>
                    <div className="relative h-full flex flex-row place-content-between items-center px-2">
                      <div className="z-10 text-white text-base text-left flex flex-row gap-2">
                        <div className="text-white">{customer.flag}</div>
                        <div className="text-white">{customer.language}</div>
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

export default LanguagesCard;