
export const displayPublishDate = (dateString: string) => {
  if (!dateString) return 'No publish date available';
  try {
    const date = new Date(dateString);
    const months = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];
    const month = months[date.getUTCMonth()];
    const day = date.getUTCDate();
    const year = date.getUTCFullYear();
    return `${month} ${day}, ${year}`;
  } catch (error) {
    console.log('Error formatting publish date: ', error);
    return dateString;
  }
};
