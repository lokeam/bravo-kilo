import { useFocusContext } from '../FocusProvider/FocusProvider';
import { Link } from 'react-router-dom';
import { FaSearchPlus } from 'react-icons/fa';
import { MdFormatListBulletedAdd } from "react-icons/md";

function EmptyLibraryCard() {
  const { searchFocusRef, addManualRef } = useFocusContext();

  const handleSearchClick = () => {
    if (searchFocusRef.current) {
      searchFocusRef.current?.focus();
    }
  };

  const handleAddManualClick = () => {
    if (addManualRef.current) {
      addManualRef.current?.focus();
    }
  }

  const sharedBtnClasses = "bg-white dark:bg-dark-tone-ink flex flex-row content-start cursor-pointer items-center px-8 py-4 mb-4 gap-4 w-full text-charcoal dark:text-white rounded-lg transition duration-500 ease-in-out border border-cadet-gray dark:border-gray-700/60 hover:border-vivid-blue dark:hover:border-vivid-blue dark:border";
  const sharedIconClasses = "relative inline-flex items-center justify-center overflow-hidden h-12 w-12 rounded-full";
  const sharedBtnTextClasses = "flex flex-col text-center content-center justify-center font-semibold";

  return(
    <div className="relative box-border flex flex-col rounded-lg min-h-min items-center content-center h-full">

    <div className="flex flex-col w-full items-center content-center">
      <div className="p-5">
        <h2 className="text-charcoal dark:text-white text-4xl font-bold pb-6">Welcome to your Library!</h2>
        <p className="text-charcoal dark:text-white pb-6">This is your new, all singing, all dancing library of meta data.</p>
        <p className="text-charcoal dark:text-white">Here's how you get started:</p>

        <button
          className={`${sharedBtnClasses} mt-4`}
          onClick={handleSearchClick}
        >
          <div className={`${sharedIconClasses} bg-lime-green`}>
            <FaSearchPlus color="#0c0c0c" size={23} />
          </div>
          <div className={`${sharedBtnTextClasses}`}>Add some books by search</div>
        </button>

        <button
          className={`${sharedBtnClasses}`}
          onClick={handleAddManualClick}
        >
          <div className={`${sharedIconClasses} bg-strong-violet`}>
            <MdFormatListBulletedAdd color="#fff" size={28} />
          </div>
          <div className={`${sharedBtnTextClasses}`}>Add some books by manual entry</div>
        </button>
        <p>For more info, visit the <Link className=" transition duration-500 ease-in-out text-strong-violet dark:text-lime-green hover:text-vivid-blue dark:hover:text-vivid-blue" to={"/support"}>Getting Started Guide</Link>.</p>
      </div>
    </div>
  </div>
  )
}

export default EmptyLibraryCard;
