import ChartCardHeader from './ChartCardHeader';
import { BookStatObj } from './BarChartCard';
import { FaTag } from "react-icons/fa";
import { FaTags } from "react-icons/fa";

interface TagsCardProps {
  userTags: BookStatObj[];
}

function TableCard({ userTags = [] }: TagsCardProps) {

  return (
    <div className="tags_card col-span-full xl:col-span-6 bg-maastricht shadow-sm rounded-xl max-h-[465px]">
      <ChartCardHeader
        hasSubHeaderBg
        topic="Tags"
      />
        <ul className="p-3">
          { userTags && userTags.length > 0 ? userTags.map((userTag, index) => {
            const hasManyTags = userTag.count > 5;

            return (
            <li
              className="flex px-2 border-b border-gray-700/60"
              key={`${index}-${userTag.label}-${userTag.count}`}>
                <div className={`flex flex-col content-center items-center justify-center w-9 h-9 rounded-full shrink-0 ${hasManyTags ? 'bg-green-600' : 'bg-red-500'} my-2 mr-3`}>
                { hasManyTags ?
                    <FaTags
                      className="bg-transparent"
                      color="text-white"
                      size={22}
                    /> :
                    <FaTag
                      className="bg-transparent"
                      color="text-white"size={22}
                    />
                }
                </div>
                <div className="grow flex items-center text-sm py-2">
                  <div className="grow flex justify-between">
                    <div className="self-center">
                      <a className="font-medium text-gray-800 hover:text-gray-900 dark:text-gray-100 dark:hover:text-white"href="#">{userTag.label}</a>
                    </div>
                    <div className="shrink-0 self-start ml-2">
                      <span className="font-medium text-gray-800 dark:text-gray-100">{userTag.count}</span>
                    </div>
                  </div>
                </div>
            </li>
            )}) : (
            <li className="flex px-2 border-b border-gray-700/60">Add some books to your library to start adding tags</li>
            )}
        </ul>
  </div>
  )
}

export default TableCard;
