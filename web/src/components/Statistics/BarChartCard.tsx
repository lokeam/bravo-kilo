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
      <div className="genre_card col-span-full mdTablet:col-span-4 bg-maastricht shadow-sm rounded-xl max-h-[465px]">
        <ChartCardHeader topic="Genre" />
        <BarChartCardBody bookData={props.booksByGenre} barColor="bg-hepatica-lt/[0.2]" totalBooks={props.totalBooks} />
      </div>
    );
  }

  if (isBarChartCardLanguages(props)) {
    return (
      <div className="language_card col-span-full lgMobile:col-span-6 mdTablet:col-span-4 bg-maastricht shadow-sm rounded-xl">
        <ChartCardHeader topic="Language" />
        <BarChartCardBody bookData={props.booksByLang} barColor="bg-maya-blue/[0.2]" totalBooks={props.totalBooks} />
      </div>
    )
  }

  return null;
}
