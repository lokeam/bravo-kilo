import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import { IoClose } from 'react-icons/io5';
import { IoAddOutline } from 'react-icons/io5';
import { IoIosWarning } from "react-icons/io";
import { MdDeleteForever } from "react-icons/md";
import { RiFileCopyLine } from "react-icons/ri";

const Settings = () => {
  // Delete Modal state
  const [opened, setOpened] = useState(false);

  // AI Preview + Modal state
  const { bookID } = useParams();
  const navigate = useNavigate();

  return (
    <section className="bg-black relative flex flex-col items-center px-5 antialiased mdTablet:pr-5 mdTablet:ml-24 h-screen">
      <div className="text-left max-w-screen-mdTablet py-24 md:pb-4 flex flex-col relative w-full">
        <h2 className="mb-4 text-3xl font-bold text-gray-900 dark:text-white">Settings</h2>

        <div className="grid gap-4 grid-cols-1 sm:gap-6 py-3">
          {/* Appearance */}
          <div className="grid w-full gap-6 lgMobile:grid-cols-2 mdTablet:col-span-1 py-5 border-t border-gray-600">
            <div className="flex flex-col mb-2 text-base font-medium text-gray-900 dark:text-white">
              <h3 className="mb-1 text-xl font-bold">Appearance</h3>
              <p className="text-nevada-gray">Customize how Bravo Kilo looks</p>
            </div>
            <div className="grid w-full">
              <button className="h-11 justify-stretch">Use my system setting</button>
            </div>
          </div>

          {/* Library Data */}
          <div className="grid w-full gap-6 lgMobile:grid-cols-2 mdTablet:col-span-1 py-5 border-t border-gray-600">
            <div className="block mb-2 text-base font-medium text-gray-900 dark:text-white">
              <h3 className="mb-1 text-xl font-bold">Data Export</h3>
            </div>
            <div className="grid w-full">
              <button className="h-11 justify-stretch">Export library data as CSV</button>
            </div>
          </div>

          {/* Cowbell */}
          <div className="grid w-full gap-6 lgMobile:grid-cols-2 mdTablet:col-span-1 py-5 border-t border-gray-600">
            <div className="block mb-2 text-base font-medium text-gray-900 dark:text-white">
              <h3 className="mb-1 text-xl font-bold">Animation</h3>
              <p className="text-nevada-gray">Enable animations and transitions</p>
            </div>
            <label className="justify-center items-center cursor-pointer grid w-full">
              <input type="checkbox" value="" className="sr-only peer" />
              <div className="relative w-14 h-7 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-majorelle rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:start-[4px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-6 after:w-6 after:transition-all dark:border-gray-600 peer-checked:bg-majorelle"></div>
            </label>
          </div>

          {/* Delete Account */}
          <div className="grid w-full gap-6 lgMobile:grid-cols-2 mdTablet:col-span-1 py-5 border-y border-gray-600">
            <div className="block mb-2 text-base font-medium text-gray-900 dark:text-white">
              <h3 className="mb-1 text-xl font-bold">Delete Account</h3>
              <p className="text-nevada-gray">Permanetly delete the account and all library data</p>
            </div>
            <div className="grid w-full">
              <button className="h-11 border-red-500 text-red-500 hover:text-white hover:bg-red-600 hover:border-red-600 focus:ring-red-900">Delete my account</button>
            </div>
          </div>
        </div>

      </div>
    </section>
  );
};

export default Settings;
