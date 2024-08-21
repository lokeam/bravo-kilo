import React from 'react';
import DoughnutChart from '../Chart/DonutChart';

const FormatsCard: React.FC = () => {
  const chartData = {
    labels: ['Physical Books', 'eBooks', 'Audio Book'],
    datasets: [
      {
        data: [34, 23, 23],
        backgroundColor: ['#7C3AED', '#0EA5E9', '#4C1D95'], // Violet and Sky colors
        hoverBackgroundColor: ['#6D28D9', '#0284C7', '#3F0D87'],
        borderWidth: 0,
      },
    ],
  };

  return (
    <div className="flex flex-col col-span-full sm:col-span-6 xl:col-span-4 bg-maastricht shadow-sm rounded-xl">
      <DoughnutChart data={chartData} width={389} height={260} />
    </div>
  );
};

export default FormatsCard;
