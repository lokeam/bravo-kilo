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

  return(
    <div className="relative box-border flex flex-col rounded-lg min-h-min items-center content-center h-full">

    <div className="flex flex-col w-full items-center content-center">
      <div className="p-5">
        <h2 className="text-4xl font-bold pb-6">Welcome to your Library!</h2>
        <p className="pb-6">This is your new, all singing, all dancing library of meta data.</p>
        <p>Here's how you get started:</p>

        <button
          className="bg-maastricht flex flex-row cursor-pointer items-center content-start rounded-lg gap-4 border border-maastricht hover:border-majorelle px-8 py-4 mt-4 mb-4 w-full"
          onClick={handleSearchClick}
        >
          <div className="relative inline-flex items-center justify-center overflow-hidden bg-margorelle-comp1-g h-12 w-12 rounded-full">
            <FaSearchPlus color="#0c0c0c" size={23} />
          </div>
          <div className="flex flex-col h-full text-center content-center justify-center font-semibold">Add some books by search</div>
        </button>

        <button
          className="bg-maastricht flex flex-row cursor-pointer items-center content-start rounded-lg gap-4 border border-maastricht hover:border-majorelle px-8 py-4 mb-4 w-full"
          onClick={handleAddManualClick}
        >
          <div className="relative inline-flex items-center justify-center overflow-hidden bg-margorelle-comp1-r h-12 w-12 rounded-full">
            <MdFormatListBulletedAdd color="#0c0c0c" size={28} />
          </div>
          <div className="flex flex-col text-center content-center justify-center font-semibold">Add some books by manual entry</div>
        </button>
        <p>For more info, visit the <Link className="" to={"/support"}>Getting Started Guide</Link>.</p>
      </div>
    </div>
  </div>
  )
}

export default EmptyLibraryCard;