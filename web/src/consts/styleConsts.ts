// Tailwind Form classes
export const TAILWIND_FORM_CLASSES = {
  LABEL: 'block mb-2 text-base font-medium text-gray-900 dark:text-white',
  LABEL_ASTERISK: 'text-red-600 ml-px',
  INPUT: 'bg-white dark:bg-black border border-gray-300 hover:border-gray-500 dark:hover:border-gray-300 dark:border-gray-600 text-gray-900 shadow-sm text-base rounded-lg focus:ring-primary-600 focus:border-primary-600 transition duration-500 block w-full p-2.5 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500 mb-1',
  ERROR: 'text-red-500',
  ERROR_BORDER: 'border-red-500 dark:border-red-500',
  ONE_COL_WRAPPER: 'block col-span-2',
  TWO_COL_WRAPPER: 'block col-span-2 mdTablet:col-span-1',
  FIELD_ARR_WRAPPER: 'bg-white-smoke border border-gray-300 dark:border-gray-600 rounded-lg p-4 mb-1 dark:bg-night',
  FIELD_ARR_COL_WRAPPER: 'flex flex-col sm:gap-6',
  FIELD_ARR_ROW_WRAPPER: 'flex w-full items-center mb-1 col-span-2',
  ADD_BUTTON: 'flex flex-row justify-between items-center dark:border-2 border-gray-200 bg-white text-charcoal shadow-sm hover:bg-gray-300 hover:text-vivid-blue-d hover:shadow-md hover:border-1 dark:hover:bg-gray-600 transition duration-500 ease-in-out border-cadet-gray hover:border-vivid-blue dark:hover:border-vivid-blue dark:bg-dark-clay dark:text-white dark:border-gray-600 py-2 px-3 mt-2',
  REMOVE_BUTTON: 'bg-white text-charcoal dark:text-cadet-gray border dark:border-2 border-gray-200 dark:border-gray-600 flex flex-row justify-between items-center shadow-sm dark:bg-dark-clay ml-4 transition duration-500 hover:bg-gray-300 dark:hover:bg-gray-600 hover:border-vivid-blue dark:hover:border-vivid-blue',
  TEXT_AREA: 'bg-white dark:bg-black block p-2.5 w-full text-base text-gray-900 rounded-lg transition duration-500 border border-gray-300 dark:border-gray-600 focus:ring-primary-500 hover:border-gray-500 focus:border-primary-500 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500 mb-1',
} as const;

export const TAILWIND_CARD_LIST_ITEM_CLASSES = {
  BORDER: 'border-b ',
  LAST_ITEM: 'border-none',
} as const;

