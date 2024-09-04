import { MdMenuBook } from "react-icons/md";

interface ImagePlaceholderProps {
  isBookDetail?: boolean;
}

const ImagePlaceholder = ({ isBookDetail }: ImagePlaceholderProps) => (
  <div className={`flex ${isBookDetail ? 'h-52 w-52' : 'w-16 h-16'} flex-col items-center justify-center rounded bg-dark-gunmetal`}>
    {
      isBookDetail ? (
        <>
          <MdMenuBook
            className="mb-3"
            size={60}
          />
          <span>No book cover image available</span>
        </>
      ) : <MdMenuBook size={36}/>
    }
  </div>
);

export default ImagePlaceholder;
