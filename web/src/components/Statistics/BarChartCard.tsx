import ChartCardHeader from './ChartCardHeader';
import BarChartCardBody from './BarChartCardBody';

export interface BookStatObj {
  label: string;
  count: number;
}

interface BarChartCardGenres {
  booksByGenre: BookStatObj[];
  totalBooks: number;
}

interface BarChartCardLanguages {
  booksByLang: BookStatObj[];
  totalBooks: number;
}

type BarChartCardProps = BarChartCardGenres | BarChartCardLanguages;


// Type Guard Checks
function isBarChartCardGenres(props: BarChartCardProps): props is BarChartCardGenres {
  return Array.isArray((props as BarChartCardGenres).booksByGenre);
}

function isBarChartCardLanguages(props: BarChartCardProps): props is BarChartCardLanguages {
  return 'booksByLang' in props;
}

export default function BarChartCard(props: BarChartCardProps) {

  if (isBarChartCardGenres(props)) {
    return (
      <div className="genre_card bg-white col-span-full mdTablet:col-span-4 shadow-xl rounded-xl max-h-[465px] dark:bg-maastricht">
        <ChartCardHeader topic="Genre" />
        <BarChartCardBody
          bookData={props.booksByGenre} barColor="bg-vivid-blue/[0.6]"
          totalBooks={props.totalBooks}
        />
      </div>
    );
  }

  if (isBarChartCardLanguages(props)) {
    return (
      <div className="language_card bg-white col-span-full lgMobile:col-span-6 mdTablet:col-span-4  shadow-xl rounded-xl dark:bg-maastricht">
        <ChartCardHeader topic="Language" />
        <BarChartCardBody
          barColor="bg-lime-green/[0.6]"
          bookData={props.booksByLang}
          isLanguageCard
          totalBooks={props.totalBooks}
        />
      </div>
    )
  }

  return null;
}
