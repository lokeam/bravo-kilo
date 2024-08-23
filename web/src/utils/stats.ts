export const convertToPercent = (num: number, total: number) => {
  return `${Math.round((num / total) * 100)}%`
};