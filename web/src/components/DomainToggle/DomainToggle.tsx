import { useState, useMemo } from "react";
import { useDomainStore } from "../../store/useDomainStore";
import { DomainType } from "../../store/useDomainStore";

import { useAuth } from "../AuthContext";
import { GiOpenBook } from "react-icons/gi";
// import { IoGameController } from "react-icons/io5";
// import { IoMdMusicalNotes } from "react-icons/io";
// import { GiFilmStrip } from "react-icons/gi";

import {Menu} from "../Menu/Menu";

export default function DomainToggle() {
  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);
  const [open, setOpen] = useState(false);

  // Get currently active domain
  const { currentDomain, setCurrentDomain, activeDomains } = useDomainStore();

  console.log('Current domain: ', currentDomain);
  console.log('Active domains: ', activeDomains);

  const handleButtonClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    // If menu is open, clicking the button again closes it
    if (open) {
      setOpen(false);
      setAnchorEl(null);
    } else {
      setAnchorEl(event.currentTarget);
      setOpen(true);
    }
  };

  const handleClose = () => {
    setOpen(false);
    setAnchorEl(null);
  };

  const handleDomainSelection = (domain: DomainType) => {
    console.log(`handleDomainSelection: ${domain}`);
    setCurrentDomain(domain);
    handleClose();
  }


  // Only render active domains
  const items = useMemo(() => {
    const domainConfig = [
      { domain: 'books' as const, label: 'Books' },
      { domain: 'games' as const, label: 'Games' },
      { domain: 'movies' as const, label: 'Movies' },
      { domain: 'music' as const, label: 'Music' },
    ];

    console.log('Before filter/map - currentDomain: ', currentDomain);

    const filteredAndMapped = domainConfig
      .filter(menuItem => {
        const isIncluded = activeDomains.includes(menuItem.domain);
        console.log(`Filtering ${menuItem.label}: included =${isIncluded}`);
        return isIncluded;
      })
      .map(item => {
        const isActive = currentDomain === item.domain;
        console.log(`Mapping ${item.label}: isActive = ${isActive}`);
        return {
          label: item.label,
          onClick: () => handleDomainSelection(item.domain),
          isActive: isActive,
        };
      });

      console.log('FInal menu items: ', filteredAndMapped);

    return filteredAndMapped;
  }, [activeDomains, currentDomain, handleDomainSelection])

  // const items = [
  //   {
  //     label: 'Books',
  //     onClick: () => handleDomainSelection('books'),
  //     isActive: currentDomain === 'books',
  //     disabled: !activeDomains.includes('books'),
  //   },
  //   {
  //     label: 'Games',
  //     onClick: () => handleDomainSelection('games'),
  //     isActive: currentDomain === 'games',
  //     disabled: !activeDomains.includes('games'),
  //   },
  //   {
  //     label: 'Movies',
  //     onClick: () => handleDomainSelection('movies'),
  //     isActive: currentDomain === 'movies',
  //     disabled: !activeDomains.includes('movies'),
  //   },
  //   {
  //     label: 'Music',
  //     onClick: () => handleDomainSelection('movies'),
  //     isActive: currentDomain === 'movies',
  //     disabled: !activeDomains.includes('movies'),
  //    },
  // ];

  console.log('Domain Toggle component');
  const { user } = useAuth();

  console.log('User from autAuth: ', user);
// relative inline-flex items-center justify-center overflow-hidden bg-gray-300 dark:bg-gray-600 h-10 w-10 rounded
// border-charcoal dark:border-gray-700/60 dark:border-2 transition duration-500 ease-in-out hover:border-vivid-blue dark:hover:border-vivid-blue bg-transparent
  return (
    <div className="flex flex-row justify-center space-x-5 rounded avatar text-dark-ebony dark:text-white">
        <button
          className="border-charcoal dark:border-gray-700/60 dark:border-2 transition duration-500 ease-in-out hover:border-vivid-blue dark:hover:border-vivid-blue bg-transparent p-2 mx-2 rounded"
          onClick={handleButtonClick}
          >
          <GiOpenBook size={25} className="text-black dark:text-az-white"/>
        </button>
        <Menu
          anchorEl={anchorEl}
          onClose={handleClose}
          open={open}
          items={items}
        />
    </div>
  )
}
