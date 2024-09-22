import PieChartCardFooter from './DonutChartCardFooter';
import DoughnutChart from '../Chart/DonutChart';
import ErrorCard from './ErrorCard';
import { BookStatObj } from './BarChartCard';

type DonutChartCardProps = {
  bookFormats: BookStatObj[];
}

function DonutChartCard({ bookFormats = [] }: DonutChartCardProps) {

  const formatCountArr = bookFormats.map((book) => book.count);
  const chartData = {
    labels: ['Physical Books', 'eBooks', 'Audio Book'],
    datasets: [
      {
        data: formatCountArr,
        backgroundColor: ['#5bf563', '#6a00b9', '#086fe8'],
        hoverBackgroundColor: ['#5bf563', '#6a00b9', '#086fe8'],
        borderWidth: 0,
      },
    ],
  };

  console.log('formatCountArr: ', formatCountArr);

  return(
    <div className="books_format_card_wrapper bg-white flex flex-col col-span-full lgMobile:col-span-6 mdTablet:col-span-4 shadow-xl rounded-xl dark:bg-maastricht dark:border-none ">

      { bookFormats && bookFormats.length > 0 ? (
        <div className="">
          <header className="books_format_header border-b border-gray-100 px-5 py-4 dark:border-gray-700/60">
            <h2 className="text-left text-charcoal text-lg font-semibold dark:text-white">Books By Format</h2>
          </header>
          <div className="bg-white flex flex-col col-span-full sm:col-span-6 xl:col-span-4 dark:bg-maastricht">
            <DoughnutChart
              data={chartData}
              height={200}
              width={200}
            />
          </div>
          <PieChartCardFooter bookFormats={bookFormats} />
        </div>
      ) : (
        <ErrorCard />
      )}
    </div>
  );
}

export default DonutChartCard;
