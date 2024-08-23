import { BookStatObj } from "./BarChartCard";

interface DonutChartHeaderProps {
  bookFormats: BookStatObj[];
}

const DonutChartCardHeader = ({ bookFormats = [] }: DonutChartHeaderProps) => {

  return (
    <>
      <header className="books_format_header border-b border-gray-700/60 px-5 py-4">
        <h2 className="text-left text-lg font-semibold">Books By Format</h2>
      </header>

      <div className="books_format_label_container flex flex-wrap px-5 py-6">
        { bookFormats && bookFormats.length > 0 ? bookFormats.map((bookFormat, index) => (
          <div
            key={`${index}-${bookFormat.label}-${bookFormat.count}`}
            className="flex flex-col items-center min-w-[33%] py-2"
          >
            <div className={`text-4xl text-center ${index !== 2 ? 'border-r' : ''} border-gray-600 w-full font-bold text-gray-800 dark:text-gray-100 mr-2 mb-1`}>
              <span>{bookFormat.count}</span>
            </div>
            <div className="text-sm text-gray-500 dark:text-gray-400 text-upper font-semibold">
              {bookFormat.label}
            </div>
          </div>
        )) : null }
      </div>
    </>
  );
}

export default DonutChartCardHeader;
