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

interface BarChartCardAuthors {
  booksByAuthor: BookStatObj[];
  totalBooks: number;
}

type BarChartCardProps = BarChartCardGenres | BarChartCardLanguages | BarChartCardAuthors;

// Type Guard Checks
function isBarChartCardGenres(props: BarChartCardProps): props is BarChartCardGenres {
  return Array.isArray((props as BarChartCardGenres).booksByGenre);
}

function isBarChartCardLanguages(props: BarChartCardProps): props is BarChartCardLanguages {
  return 'booksByLang' in props;
}

function isBarChartCardAuthors(props: BarChartCardProps): props is BarChartCardAuthors {
  return Array.isArray((props as BarChartCardAuthors).booksByAuthor);
}

export default function BarChartCard(props: BarChartCardProps) {
  const commonClasses = "bg-white dark:bg-eight-ball box-border col-span-full mdTablet:col-span-4 rounded-xl shadow-xl  dark:border dark:border-gray-700/60 overflow-y-hidden";

  if (isBarChartCardGenres(props)) {
    return (
      <div className={`genre_card ${commonClasses}`}>
        <ChartCardHeader topic="Genre" />
        <BarChartCardBody
          bookData={props.booksByGenre}
          barColor="bg-vivid-blue/[0.6]"
          totalBooks={props.totalBooks}
        />
      </div>
    );
  }

  if (isBarChartCardLanguages(props)) {
    return (
      <div className={`language_card ${commonClasses}`}>
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

  if (isBarChartCardAuthors(props)) {
    return (
      <div className={`author_card ${commonClasses}`}>
        <ChartCardHeader topic="Author" />
        <BarChartCardBody
          barColor="bg-strong-violet/[0.6]"
          bookData={props.booksByAuthor}
          totalBooks={props.totalBooks}

        />
      </div>
    );
  }

  return null;
}
