import { convertToPercent } from '../../utils/stats';
import { BookStatObj } from './BarChartCard';
import { languageCodes } from '../../consts/languageCodes';

interface BarChartCardBodyProps {
  barColor: string;
  bookData: BookStatObj[];
  totalBooks: number;
  isLanguageCard?: boolean;
}

function BarChartCardBody({
  bookData = [],
  totalBooks = 0,
  barColor = "bg-margorelle-comp1-r",
  isLanguageCard = false,
}: BarChartCardBodyProps) {

  return (
    <div className="pb-3 px-3">
      <div className="overflow-x-hidden">
        <ul className="text-sm divide-y leading-6 dark:divide-gray-700/60 mt-3 mb-4">
            { bookData && bookData.length > 0 ? bookData.map((book, index) => {
              const bgBarWidth = convertToPercent(book.count, totalBooks);

              return (
                <li
                  className="relative border-none mb-2 p-1"
                  key={`${index}-${book.label}-${book.count}`}
                >
                  <div
                    className={`horizontal_bar top-0 left-0 ${barColor} rounded absolute p-4`}
                    style={{width: `${bgBarWidth}`}}
                  ></div>
                  <div className="relative h-full flex flex-row place-content-between items-center px-2">
                    <div className="z-10 text-white text-base text-left flex flex-row gap-2">
                      <div className="text-black dark:text-white capitalize">
                        { isLanguageCard ? languageCodes[book.label] : book.label.toLowerCase() }
                      </div>
                    </div>
                    <div className="text-black text-right dark:text-white">{book.count}&nbsp; /&nbsp; {totalBooks}</div>
                  </div>
                </li>
              )
            }) : (
                <li>It appears as if you haven't saved any books to your library yet</li>
              )
            }
        </ul>
      </div>
    </div>
  );
}

export default BarChartCardBody;
