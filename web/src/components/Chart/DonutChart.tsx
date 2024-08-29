import React, { useRef, useEffect } from 'react';
import { Chart, DoughnutController, ArcElement, Tooltip } from 'chart.js';

Chart.register(DoughnutController, ArcElement, Tooltip);

interface DoughnutChartProps {
  data: any;
  width: number;
  height: number;
}

const DoughnutChart: React.FC<DoughnutChartProps> = ({ data, width, height }) => {
  const canvas = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (!canvas.current) return;

    const ctx = canvas.current.getContext('2d');
    if (!ctx) return;

    const newChart = new Chart(ctx, {
      type: 'doughnut',
      data: data,
      options: {
        cutout: '80%',
        plugins: {
          legend: { display: false },
          tooltip: {
            titleColor: '#333',
            bodyColor: '#666',
            backgroundColor: '#fff',
            borderColor: '#ddd',
          },
        },
        maintainAspectRatio: false,
      },
    });

    return () => newChart.destroy();
  }, [data]);

  return (
    <div className="flex justify-center px-5">
      <canvas ref={canvas} width={width} height={height}></canvas>
    </div>
  );
};

export default DoughnutChart;