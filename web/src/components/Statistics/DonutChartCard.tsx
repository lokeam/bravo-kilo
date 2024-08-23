import PieChartCardHeader from './DonutChartCardHeader';
import DoughnutChart from '../Chart/DonutChart';
import ErrorCard from './ErrorCard';
import { BookStatObj } from './BarChartCard';

interface DonutChartCardProps {
  bookFormats: BookStatObj[];
}

const DonutChartCard = ({ bookFormats = [] }: DonutChartCardProps) => {

  const formatCountArr = bookFormats.map((book) => book.count);
  const chartData = {
    labels: ['Physical Books', 'eBooks', 'Audio Book'],
    datasets: [
      {
        data: formatCountArr,
        backgroundColor: ['#7C3AED', '#0EA5E9', '#4C1D95'], // Violet and Sky colors
        hoverBackgroundColor: ['#6D28D9', '#0284C7', '#3F0D87'],
        borderWidth: 0,
      },
    ],
  };

  return(
    <div className="books_format_card_wrapper flex flex-col col-span-full lgMobile:col-span-6 mdTablet:col-span-4 bg-maastricht shadow-sm rounded-xl">
      <PieChartCardHeader bookFormats={bookFormats} />

      { bookFormats && bookFormats.length > 0 ? (
        <div className="flex-grow pb-4">
          <div className="flex flex-col col-span-full sm:col-span-6 xl:col-span-4 bg-maastricht shadow-sm rounded-xl">
            <DoughnutChart data={chartData} width={389} height={260} />
          </div>
        </div>
      ) : (
        <ErrorCard />
      )}
    </div>
  );
}

export default DonutChartCard;
