import React, { useRef, useEffect } from 'react';
import { Chart, DoughnutController, ArcElement, Tooltip } from 'chart.js';
import { useThemeStore } from '../../store/useThemeStore';

Chart.register(DoughnutController, ArcElement, Tooltip);

interface DoughnutChartProps {
  data: any;
  width: number;
  height: number;
  centerText?: string;
}

const DoughnutChart: React.FC<DoughnutChartProps> = ({ data, width, height, centerText }) => {
  const canvas = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const { theme } = useThemeStore();
  const isDarkMode = theme === 'dark';

  useEffect(() => {
    if (!canvas.current || !containerRef.current) return;

    const ctx = canvas.current.getContext('2d');
    if (!ctx) return;

    const centerTextPlugin = {
      id: 'centerText',
      afterDraw: (chart: any) => {
        if (centerText) {
          const { ctx, chartArea: { left, top, right, bottom } } = chart;
          const centerX = (left + right) / 2;
          const centerY = (top + bottom) / 2;

          const [ firstLine, secondLine ] = centerText.split(' ');

          ctx.save();
          ctx.textAlign = 'center';
          ctx.textBaseline = 'middle';

          // First line
          ctx.fillStyle = isDarkMode ? 'rgb(243, 244, 246)' : 'rgb(31, 41, 55)';
          ctx.font = 'bold 36px system-ui';
          ctx.fillText(firstLine, centerX, centerY - 10);

          // Second line
          ctx.fillStyle = 'rgb(107, 114, 128)';
          ctx.font = 'bold 16px system-ui';
          ctx.fillText(secondLine, centerX, centerY + 20);
          ctx.restore();
        }
      }
    };

    const newChart = new Chart(ctx, {
      type: 'doughnut',
      data: data,
      options: {
        cutout: '80%',
        plugins: {
          legend: {
            display: true,
            position: 'bottom',
            maxHeight: 50,
            maxWidth: 200,
          },
          tooltip: {
            titleColor: '#fff',
            bodyColor: '#121212',
            backgroundColor: '#121212',
            borderColor: '#374151',
            borderWidth: 1,
          },
        },
        responsive: true,
        maintainAspectRatio: false,
        hover: {
          mode: 'nearest',
          intersect: true,
        },
      },
      plugins: [centerTextPlugin],
    });

    return () => newChart.destroy();
  }, [data, centerText]);

  return (
    <div
      className="flex justify-center p-6 text-charcoal dark:text-white"
      ref={containerRef}
      style={{
        '--first-line-color': 'rgb(243, 244, 246)',
        '--second-line-color': 'rgb(156, 163, 175)',
      } as React.CSSProperties}
    >
      <canvas
        ref={canvas}
        width={width}
        height={height}
      ></canvas>
    </div>
  );
};

export default DoughnutChart;
