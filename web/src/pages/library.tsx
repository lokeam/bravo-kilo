import { useState } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useAuth } from "../components/AuthContext";
import axios from "axios";

import TopNavigation from "../components/TopNav/TopNav";
import SideNavigation from "../components/SideNav/SideNavigation";
import CardList from "../components/CardList/CardList";
import Modal from '../components/Modal/Modal';
import '../components/Modal/Modal.css';


import { PiArrowsDownUp } from "react-icons/pi";

export interface Book {
  authors: string[];
  createdAt: string;
  description: string;
  genres: string[];
  id: number;
  imageLinks: string[];
  language: string;
  pageCount: number;
  publishDate: string;
  subtitle?: string;
  title: string;
}

const fetchUserBooks = async (): Promise<Book[]> => {
  const { data } = await axios.get(`${import.meta.env.VITE_API_ENDPOINT}/api/v1/user/books`, {
    withCredentials: true
  });
  return data.books;
};

const Library = () => {
  const [opened, setOpened] = useState(false);

  const { logout } = useAuth();
  const { search } = useLocation();
  const query = new URLSearchParams(search);
  const userID = parseInt(query.get('userID') || '0', 10);

  const [sortCriteria, setSortCriteria] = useState<"title" | "publishDate" | "author">("title");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");

  const { data: books, isLoading, isError } = useQuery({
    queryKey: ['userBooks'],
    queryFn: fetchUserBooks,
    enabled: !!userID,
  });

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (isError) {
    return <div>Error loading books</div>;
  }


  const handleSort = (criteria: "title" | "publishDate" | "author" | "pageCount") => {
    setSortCriteria(criteria);
    setSortOrder(prevOrder => prevOrder === 'asc' ? 'desc' : 'asc');
    setOpened(false);
  };

  const sortedBooks = books?.slice().sort((a, b) => {
    if (sortCriteria === "title") {
      return sortOrder === "asc" ? a.title.localeCompare(b.title) : b.title.localeCompare(a.title);
    } else if (sortCriteria === "publishDate") {
      return sortOrder === "asc"
        ? new Date(a.publishDate).getTime() - new Date(b.publishDate).getTime()
        : new Date(b.publishDate).getTime() - new Date(a.publishDate).getTime();
    } else if (sortCriteria === "author") {
      const aSurname = a.authors[0].split(" ").pop() || "";
      const bSurname = b.authors[0].split(" ").pop() || "";
      return sortOrder === "asc" ? aSurname.localeCompare(bSurname) : bSurname.localeCompare(aSurname);
    } else {
      return sortOrder === "asc" ? a.pageCount - b.pageCount : b.pageCount - a.pageCount;
    }
  });

  const sortButtonTitle = {
    'title': 'Title: A to Z',
    'author': 'Author: A to Z',
    'publishDate': 'Release date: New to Old',
    'pageCount': 'Page count: Short to Long',
  };

  console.log('books: ', sortedBooks)

  const openModal = () => setOpened(true);
  const closeModal = () => setOpened(false);

  return (
    <div className="bk_lib flex flex-col items-center place-content-around p-6 lg:px-8 antialiased md:ml-28 h-screen pt-24">
      <TopNavigation />
      <SideNavigation />

      <h1>Library</h1>
      <button onClick={logout}>Sign out of your Kilo Bravo account</button>

      <Modal opened={opened} onClose={closeModal} title="">
        <button onClick={() => handleSort("publishDate")} className="flex flex-row bg-transparent mr-1">
          Release date: New to Old
        </button>
        <button onClick={() => handleSort("pageCount")} className="flex flex-row bg-transparent mr-1">
          Page count: Short to Long
        </button>
        <button onClick={() => handleSort("title")} className="flex flex-row bg-transparent mr-1">
          Title: A to Z
        </button>
        <button onClick={() => handleSort("author")} className="flex flex-row bg-transparent mr-1">
          Author: A to Z
        </button>
      </Modal>

      <div className="flex flex-row relative w-full max-w-7xl justify-between items-center text-left text-white border-b-2 border-solid border-zinc-700 pb-6 mb-2">
        <div className="mt-1">{sortedBooks.length} volumes</div>

        <div className="flex flex-row">
          <button className="flex flex-row justify-between" onClick={openModal}>
            <PiArrowsDownUp className="w-5 h-5 pt-1 mr-2" color="white"/>
            <span>{sortButtonTitle[sortCriteria]}</span>
          </button>
        </div>

      </div>
      {sortedBooks && <CardList books={sortedBooks} />}

      <Outlet />
    </div>
  )
}

export default Library;
