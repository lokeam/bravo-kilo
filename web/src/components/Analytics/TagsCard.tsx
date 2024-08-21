

const TagsCard = () => {
  return (
    <div className="col-span-full xl:col-span-6 bg-maastricht shadow-sm rounded-xl">
    <header className="px-5 py-4 border-b border-gray-100 dark:border-gray-700/60">
        <h2 className="text-left font-semibold text-gray-800 dark:text-gray-100">Personal Tags</h2></header>
    <div className="p-3">
        <div>
          <ul className="flex place-content-between text-xs uppercase text-gray-400 bg-midnight-navy rounded bg-opacity-50 p-3">
              <li className="font-semibold">Tag Name</li>
              <li className="font-semibold">Tag Count</li>
            </ul>
            <ul className="my-1">
                <li className="flex px-2 border-b border-gray-700/60">
                    <div className="w-9 h-9 rounded-full shrink-0 bg-red-500 my-2 mr-3">
                        <svg className="w-9 h-9 fill-current text-white" viewBox="0 0 36 36">
                            <path d="M17.7 24.7l1.4-1.4-4.3-4.3H25v-2H14.8l4.3-4.3-1.4-1.4L11 18z"></path>
                        </svg>
                    </div>
                    <div className="grow flex items-center text-sm py-2">
                        <div className="grow flex justify-between">
                            <div className="self-center"><a className="font-medium text-gray-800 hover:text-gray-900 dark:text-gray-100 dark:hover:text-white" href="#0">Audible</a></div>
                            <div className="shrink-0 self-start ml-2"><span className="font-medium text-gray-800 dark:text-gray-100">12</span></div>
                        </div>
                    </div>
                </li>
                <li className="flex px-2 border-b border-gray-700/60">
                    <div className="w-9 h-9 rounded-full shrink-0 bg-green-500 my-2 mr-3">
                        <svg className="w-9 h-9 fill-current text-white" viewBox="0 0 36 36">
                            <path d="M18.3 11.3l-1.4 1.4 4.3 4.3H11v2h10.2l-4.3 4.3 1.4 1.4L25 18z"></path>
                        </svg>
                    </div>
                    <div className="grow flex items-center text-sm py-2">
                        <div className="grow flex justify-between">
                            <div className="self-center"><a className="font-medium text-gray-800 hover:text-gray-900 dark:text-gray-100 dark:hover:text-white" href="#0">BOOX Note</a></div>
                            <div className="shrink-0 self-start ml-2"><span className="font-medium text-green-600">10</span></div>
                        </div>
                    </div>
                </li>
                <li className="flex px-2 border-b border-gray-700/60">
                    <div className="w-9 h-9 rounded-full shrink-0 bg-green-500 my-2 mr-3">
                        <svg className="w-9 h-9 fill-current text-white" viewBox="0 0 36 36">
                            <path d="M18.3 11.3l-1.4 1.4 4.3 4.3H11v2h10.2l-4.3 4.3 1.4 1.4L25 18z"></path>
                        </svg>
                    </div>
                    <div className="grow flex items-center text-sm py-2">
                        <div className="grow flex justify-between">
                            <div className="self-center"><a className="font-medium text-gray-800 hover:text-gray-900 dark:text-gray-100 dark:hover:text-white" href="#0">Google Drive</a></div>
                            <div className="shrink-0 self-start ml-2"><span className="font-medium text-green-600">9</span></div>
                        </div>
                    </div>
                </li>
                <li className="flex px-2 border-b border-gray-700/60">
                    <div className="w-9 h-9 rounded-full shrink-0 bg-green-500 my-2 mr-3">
                        <svg className="w-9 h-9 fill-current text-white" viewBox="0 0 36 36">
                            <path d="M18.3 11.3l-1.4 1.4 4.3 4.3H11v2h10.2l-4.3 4.3 1.4 1.4L25 18z"></path>
                        </svg>
                    </div>
                    <div className="grow flex items-center text-sm py-2">
                        <div className="grow flex justify-between">
                            <div className="self-center"><a className="font-medium text-gray-800 hover:text-gray-900 dark:text-gray-100 dark:hover:text-white" href="#0">Kindle</a></div>
                            <div className="shrink-0 self-start ml-2"><span className="font-medium text-green-600">9</span></div>
                        </div>
                    </div>
                </li>
                <li className="flex px-2 border-b border-gray-700/60">
                    <div className="w-9 h-9 rounded-full shrink-0 bg-gray-200 my-2 mr-3">
                        <svg className="w-9 h-9 fill-current text-gray-400" viewBox="0 0 36 36">
                            <path d="M21.477 22.89l-8.368-8.367a6 6 0 008.367 8.367zm1.414-1.413a6 6 0 00-8.367-8.367l8.367 8.367zM18 26a8 8 0 110-16 8 8 0 010 16z"></path>
                        </svg>
                    </div>
                    <div className="grow flex items-center text-sm py-2">
                        <div className="grow flex justify-between">
                            <div className="self-center"><a className="font-medium text-gray-800 hover:text-gray-900 dark:text-gray-100 dark:hover:text-white" href="#0">Rakuten Kobo</a> Market Ltd 70 Wilson St London</div>
                            <div className="shrink-0 self-start ml-2"><span className="font-medium text-gray-800 dark:text-gray-100">7</span></div>
                        </div>
                    </div>
                </li>
                <li className="flex px-2 border-b border-gray-700/60">
                    <div className="w-9 h-9 rounded-full shrink-0 bg-red-500 my-2 mr-3">
                        <svg className="w-9 h-9 fill-current text-white" viewBox="0 0 36 36">
                            <path d="M17.7 24.7l1.4-1.4-4.3-4.3H25v-2H14.8l4.3-4.3-1.4-1.4L11 18z"></path>
                        </svg>
                    </div>
                    <div className="grow flex items-center text-sm py-2">
                        <div className="grow flex justify-between">
                            <div className="self-center"><a className="font-medium text-gray-800 hover:text-gray-900 dark:text-gray-100 dark:hover:text-white" href="#0">Study top bookshelf</a></div>
                            <div className="shrink-0 self-start ml-2"><span className="font-medium text-gray-800 dark:text-gray-100">3</span></div>
                        </div>
                    </div>
                </li>
            </ul>
        </div>
    </div>
  </div>
  )
}

export default TagsCard;
