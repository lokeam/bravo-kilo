import { Suspense } from 'react';
import { Route, Routes } from 'react-router-dom';
import { AppProvider } from './components/AuthContext';
import { FocusProvider } from './components/FocusProvider/FocusProvider';
import { NetworkStatusProvider } from './components/NetworkStatusProvider/NetworkStatusProvider';
import ThemeProvider from './components/ThemeProvider/ThemeProvider';
import AuthenticatedLayout from './pages/AuthLayout';
import ProtectedRoute from './components/ProtectedRoute';
import Home from './pages/Home';
import EditBook from './pages/EditBook';
import AddBookGateway from './pages/AddBookGateway';
import AddManual from './pages/AddManual';
import AddUpload from './pages/AddUpload';
import Search from './pages/Search';
import Settings from './pages/Settings';
import Login from './pages/Login';
import Library from './pages/Library';
import AuthorGenre from './pages/AuthorGenre';
import BookDetail from './pages/BookDetail';
import NotFound from './pages/NotFound';
import Loading from './components/Loading/Loading';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

import './App.css'
import MtkHome from './pages/MktHome';
import MktBlogLayout from './components/Mkt/MktBlogLayout';

function App() {
  return (
    <NetworkStatusProvider>
      <ThemeProvider>
        <AppProvider>
          <FocusProvider>
            <Suspense fallback={<Loading />}>
              <Routes>
                <Route path="/bravo-kilo" element={<MtkHome />} />
                <Route path="/login" element={<Login />}/>
                <Route path="/support" element={<MktBlogLayout />} />
                <Route element={<ProtectedRoute><AuthenticatedLayout /></ProtectedRoute>}>
                  <Route path="/home" element={<Home />} />
                  <Route path="/library" element={<Library />} />
                  <Route path="/library/:authorID" element={<AuthorGenre />} />
                  <Route path="/library/:genreID" element={<AuthorGenre /> }/>
                  <Route path="/library/books/add/gateway" element={<AddBookGateway />} />
                  <Route path="/library/books/add/manual" element={<AddManual /> }/>
                  <Route path="/library/books/add/search" element={<AddManual /> }/>
                  <Route path="/library/books/add/upload" element={<AddUpload /> }/>
                  <Route path="/library/books/search" element={<Search />} />
                  <Route path="/library/books/:bookTitle" element={<BookDetail />} />
                  <Route path="/library/books/:bookID/edit" element={<EditBook />} />
                  <Route path="/settings" element={<Settings />} />
                </Route>
                <Route path="*" element={<NotFound />}/>
              </Routes>
            </Suspense>
          </FocusProvider>
          <ReactQueryDevtools />
        </AppProvider>
      </ThemeProvider>
    </NetworkStatusProvider>
  );
}

export default App;
