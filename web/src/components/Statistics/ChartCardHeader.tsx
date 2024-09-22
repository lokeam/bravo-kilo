interface BarChartHeaderProps {
  topic: string;
  hasSubHeaderBg?: boolean;
}

function ChartCardHeader({ topic = '', hasSubHeaderBg = false }: BarChartHeaderProps) {
  const barChartHeaderText = `Books by ${topic}`;
  const tableHeaderText = `Personal ${topic}`;

  return (
    <>
      <header className="px-5 py-4 border-b border-gray-100 dark:border-gray-700/60">
        <h2 className="chart_card_header text-left text-lg font-semibold text-gray-800 dark:text-gray-100">{hasSubHeaderBg ? tableHeaderText : barChartHeaderText}</h2>
      </header>

      <ul
        className={
          `flex place-content-between text-charcoal text-xs uppercase dark:text-gray-400 bg-opacity-50 font-semibold
          ${hasSubHeaderBg ? 'bg-gray-200 rounded p-3 mt-3 mx-3 dark:bg-midnight-navy' : 'pt-3 px-3'}`}
        >
        <li>{hasSubHeaderBg ? `tag name` : topic}</li>
        <li>{`${hasSubHeaderBg ? 'tag' : 'book'} count`}</li>
      </ul>
    </>
  );
}

export default ChartCardHeader;
