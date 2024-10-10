import PieChartCardFooter from './DonutChartCardFooter';
import DoughnutChart from '../Chart/DonutChart';
import ErrorCard from './ErrorCard';
import { BookStatObj } from './BarChartCard';

type DonutChartCardProps = {
  bookFormats: BookStatObj[];
  totalBooks: number;
}

function DonutChartCard({ bookFormats = [], totalBooks = 0 }: DonutChartCardProps) {

  const formatCountArr = bookFormats.map((book) => book.count);
  const chartData = {
    labels: ['Physical Books', 'eBooks', 'Audio Books'],
    datasets: [
      {
        data: formatCountArr,
        backgroundColor: ['#5bf563', '#6a00b9', '#086fe8'],
        hoverBackgroundColor: ['#5bf563', '#6a00b9', '#086fe8'],
        borderWidth: 0,
      },
    ],
  };

  console.log('totalBooks: ', totalBooks);
  return(
    <div className="donut_card_wrapper bg-white flex flex-col col-span-full mdTablet:col-span-4 shadow-xl rounded-xl dark:bg-eight-ball dark:border dark:border-gray-700/60">
      { bookFormats && bookFormats.length > 0 ? (
        <>
          <header className="donut_card_header border-b border-gray-100 px-5 py-4 dark:border-gray-700/60">
            <h2 className="text-left text-charcoal text-lg font-semibold dark:text-white">Books By Format</h2>
          </header>
          <div className="bg-white dark:bg-eight-ball flex flex-col">
            <DoughnutChart
              data={chartData}
              height={200}
              width={200}
              centerText={`${totalBooks.toString()} Total`}
            />
          </div>
          <PieChartCardFooter bookFormats={bookFormats} />
        </>
      ) : (
        <ErrorCard />
      )}
    </div>
  );
}

export default DonutChartCard;
