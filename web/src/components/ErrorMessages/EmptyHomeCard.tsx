import { Link } from 'react-router-dom';

function EmptyHomeCard() {

  return(
    <div className="relative box-border flex flex-col rounded-lg min-h-min items-center content-center h-full">
      <div className="flex flex-col w-full items-center content-center p-5">
          <h2 className="text-4xl font-bold pb-6">Your library. Summarized.</h2>
          <p className="pb-6">Never lose track all of your books data. The Q-Ko dashboard does it automatically.</p>

          <Link
            className="bg-margorelle-d3 hover:text-white hover:bg-majorelle text-white flex flex-row cursor-pointer items-center content-center rounded-lg gap-4 px-8 py-4 mt-4 mb-4"
            to={'/add'}
          > Add some books to get started
          </Link>
          <p>For more info, visit the <Link className="" to={"/support"}>Getting Started Guide</Link>.</p>
      </div>
    </div>
  )
}

export default EmptyHomeCard;