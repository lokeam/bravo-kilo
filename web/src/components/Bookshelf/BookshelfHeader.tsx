import React from 'react';

interface BookshelfHeaderProps {
  heading?: string;
}

const BookshelfHeader = ({ heading = 'Default Heading' }: BookshelfHeaderProps) => {
  return (
    <div className="bookshelf_heading items-end justify-end mx-3">
      <div className="flex flex-1 min-w-0">
        <h2 className="text-2xl font-bold text-white inline-block max-w-full overflow-hidden text-ellipsis whitespace-nowrap select-none mb-2">
          {heading}
        </h2>
      </div>
    </div>
  )
}

export default React.memo(BookshelfHeader);
