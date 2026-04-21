'use client';

import { useState } from 'react';
import {
  Bell,
  Settings,
  User,
  ArrowUpRight,
  Play,
  Pause,
  Clock,
  CheckCircle2,
  Monitor,
  ChevronDown,
  MoreVertical,
  MessageSquare,
  Zap,
  Briefcase,
  Link as LinkIcon
} from 'lucide-react';

const MENUS = ['Dashboard', 'People', 'Hiring', 'Devices', 'Apps', 'Salary', 'Calendar', 'Reviews'];

export default function Dashboard() {
  const [activeMenu, setActiveMenu] = useState('Dashboard');

  return (
    <div className="min-h-screen bg-[#c8cbd0] p-4 flex items-center justify-center font-sans text-slate-800">
      <div className="w-full max-w-[1240px] h-[800px] bg-gradient-to-br from-[#f8f9e9] via-[#f7ebd4] to-[#f4d89a] rounded-[40px] shadow-2xl p-8 flex flex-col overflow-hidden relative">
        
        {/* HEADER */}
        <header className="flex items-center justify-between mb-10 text-sm">
          <div className="flex items-center px-6 py-2 border border-slate-300 rounded-full bg-white/50 backdrop-blur-sm">
            <span className="text-xl font-normal tracking-tight">Crextio</span>
          </div>

          <div className="flex items-center gap-1 bg-white/50 backdrop-blur-md rounded-full p-1 shadow-sm">
            {MENUS.map((menu) => (
              <button
                key={menu}
                onClick={() => setActiveMenu(menu)}
                className={`px-5 py-2 rounded-full font-medium transition-colors ${
                  activeMenu === menu
                    ? 'bg-[#292b2a] text-white shadow-md'
                    : 'text-slate-600 hover:bg-white/50'
                }`}
              >
                {menu}
              </button>
            ))}
          </div>

          <div className="flex items-center gap-3">
            <button className="flex items-center gap-2 px-5 py-2.5 bg-white/70 backdrop-blur-sm rounded-full shadow-sm font-medium hover:bg-white transition-colors text-slate-700">
              <Settings size={18} />
              Setting
            </button>
            <button className="p-3 bg-white/70 backdrop-blur-sm rounded-full shadow-sm hover:bg-white text-slate-700">
              <Bell size={18} />
            </button>
            <button className="p-3 bg-white/70 backdrop-blur-sm rounded-full shadow-sm hover:bg-white text-slate-700">
              <User size={18} />
            </button>
          </div>
        </header>

        {/* WELCOME SECTION */}
        <div className="mb-10">
          <h1 className="text-[44px] font-light mb-8 tracking-tight text-slate-900">Welcome in, Nixtio</h1>
          <div className="flex items-end justify-between">
            <div className="flex gap-4">
              <div className="space-y-3">
                <p className="text-xs font-medium text-slate-500">Interviews</p>
                <div className="px-8 py-2.5 bg-[#292b2a] text-white rounded-full text-[13px] shadow-sm">15%</div>
              </div>
              <div className="space-y-3">
                <p className="text-xs font-medium text-slate-500">Hired</p>
                <div className="px-8 py-2.5 bg-[#fad961] text-slate-900 rounded-full text-[13px] font-medium shadow-sm">15%</div>
              </div>
              <div className="space-y-3 w-56">
                <p className="text-xs font-medium text-slate-500">Project time</p>
                <div className="w-full flex items-center h-[40px] px-4 rounded-full bg-white/40 border border-white/60 overflow-hidden relative">
                  <div className="absolute inset-0 opacity-50" style={{ backgroundImage: 'repeating-linear-gradient(45deg, transparent, transparent 4px, rgba(0,0,0,0.05) 4px, rgba(0,0,0,0.05) 8px)' }}></div>
                  <span className="text-[13px] text-slate-500 relative z-10 w-full mb-0.5">60%</span>
                  <div className="absolute left-1/4 transform -translate-x-1/2 w-[1px] h-4 bg-slate-300"></div>
                  <div className="absolute left-2/4 transform -translate-x-1/2 w-[1px] h-4 bg-slate-300"></div>
                  <div className="absolute left-3/4 transform -translate-x-1/2 w-[1px] h-4 bg-slate-300"></div>
                </div>
              </div>
              <div className="space-y-3 w-28">
                <p className="text-xs font-medium text-slate-500">Output</p>
                <div className="w-full px-6 py-2.5 bg-transparent border-[1.5px] border-slate-400 rounded-full text-[13px] text-slate-700">10%</div>
              </div>
            </div>

            <div className="flex items-end gap-10 pr-4">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-[#e8e9dc] flex items-center justify-center -mr-1">
                  <User size={14} className="text-slate-600" />
                </div>
                <div className="text-[44px] font-light leading-none -tracking-wide">78</div>
                <div className="text-[11px] text-slate-500 font-medium leading-tight mt-2 mr-2">Employe</div>
              </div>
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-[#e8e9dc] flex items-center justify-center -mr-1">
                   <User size={14} className="text-slate-600" />
                </div>
                <div className="text-[44px] font-light leading-none -tracking-wide">56</div>
                <div className="text-[11px] text-slate-500 font-medium leading-tight mt-2 mr-2">Hirings</div>
              </div>
              <div className="flex items-center gap-3">
                 <div className="w-8 h-8 rounded-full bg-[#e8e9dc] flex items-center justify-center -mr-1">
                    <Monitor size={14} className="text-slate-600" />
                 </div>
                <div className="text-[44px] font-light leading-none -tracking-wide">203</div>
                <div className="text-[11px] text-slate-500 font-medium leading-tight mt-2">Projects</div>
              </div>
            </div>
          </div>
        </div>

        {/* MAIN GRID */}
        <div className="flex-1 grid grid-cols-12 gap-5 h-[340px]">
          
          {/* COL 1 */}
          <div className="col-span-3 flex flex-col gap-5">
            <div className="relative rounded-[32px] overflow-hidden h-[200px] shrink-0 shadow-[0_8px_30px_rgb(0,0,0,0.12)]">
              <img src="https://images.unsplash.com/photo-1534528741775-53994a69daeb?q=80&w=400&h=400&fit=crop" className="w-full h-full object-cover scale-105" alt="Profile" />
              <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent" />
              <div className="absolute bottom-4 left-5 right-5 flex justify-between items-end">
                <div>
                  <h3 className="text-white font-medium text-lg">Lora Piterson</h3>
                  <p className="text-white/60 text-xs">UX/UI Designer</p>
                </div>
                <div className="px-3 py-1 bg-white/20 backdrop-blur-md rounded-2xl text-white text-[13px] border border-white/20">$1,200</div>
              </div>
            </div>

            <div className="bg-white/60 backdrop-blur-md rounded-[32px] p-5 shadow-sm flex-1 flex flex-col justify-between">
              <div className="flex justify-between items-center pb-2 border-b border-slate-200/60">
                <span className="font-medium text-sm text-slate-800">Pension contributions</span>
                <ChevronDown size={14} className="text-slate-400" />
              </div>
              <div className="space-y-3 pb-2 border-b border-slate-200/60">
                <div className="flex justify-between items-center text-slate-800">
                  <span className="font-medium text-sm">Devices</span>
                  <ChevronDown size={14} className="text-slate-800 rotate-180" />
                </div>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-12 h-10 bg-gradient-to-br from-red-500 to-purple-600 rounded-[10px] flex items-end justify-center overflow-hidden relative">
                      <div className="w-10 h-6 bg-black rounded-t-sm absolute bottom-0 flex justify-center pt-0.5">
                         <div className="w-4 h-[1px] bg-slate-400"></div>
                      </div>
                    </div>
                    <div>
                      <p className="text-[13px] font-medium text-slate-800">MacBook Air</p>
                      <p className="text-[11px] text-slate-500">Version M1</p>
                    </div>
                  </div>
                  <MoreVertical size={16} className="text-slate-400" />
                </div>
              </div>
              <div className="flex justify-between items-center pb-2 border-b border-slate-200/60">
                <span className="font-medium text-sm text-slate-800">Compensation Summary</span>
                <ChevronDown size={14} className="text-slate-400" />
              </div>
              <div className="flex justify-between items-center pt-1">
                <span className="font-medium text-sm text-slate-800">Employee Benefits</span>
                <ChevronDown size={14} className="text-slate-400" />
              </div>
            </div>
          </div>

          {/* COL 2 & 3 */}
          <div className="col-span-6 flex flex-col gap-5">
            <div className="flex gap-5 h-[200px]">
              {/* Progress */}
              <div className="flex-[0.55] bg-white/70 backdrop-blur-md rounded-[32px] p-6 shadow-sm relative group">
                <div className="absolute top-5 right-5 p-1.5 rounded-full border border-slate-200 group-hover:bg-slate-50 transition-colors">
                  <ArrowUpRight size={14} className="text-slate-500" />
                </div>
                <h3 className="text-[17px] font-medium mb-1 text-slate-800">Progress</h3>
                <div className="flex items-end gap-2.5">
                  <span className="text-[28px] font-light leading-none">6.1 h</span>
                  <span className="text-[11px] text-slate-500 leading-tight w-14 mb-0.5 font-medium">Work Time this week</span>
                </div>
                
                <div className="flex items-end justify-between h-[80px] mt-4 px-1 relative">
                  <div className="absolute inset-x-0 top-1/2 border-t-[1.5px] border-dashed border-slate-200 -z-10 transform -translate-y-1/2"></div>
                  {['S','M','T','W','T','F','S'].map((day, i) => {
                    const isToday = i === 4;
                    const h = isToday ? '100%' : (i===0 || i===6) ? '30%' : (i===1 ? '60%' : i===2 ? '40%' : i===3 ? '70%' : '50%');
                    return (
                    <div key={i} className="flex flex-col items-center h-full justify-end relative group/bar">
                      <div className="absolute top-1/2 -mt-[3px] w-1.5 h-1.5 rounded-full bg-slate-200 z-0"></div>
                      <div className={`w-[3px] rounded-full z-10 transition-all ${isToday ? 'bg-[#fad961]' : i === 0 || i === 6 ? 'bg-slate-300/50' : 'bg-[#292b2a]'}`} 
                        style={{ height: h, marginBottom: '6px' }}
                      ></div>
                      <span className={`text-[10px] font-medium ${isToday ? 'text-slate-800' : 'text-slate-400'}`}>{day}</span>
                      
                      {isToday && (
                        <div className="absolute -top-6 whitespace-nowrap bg-[#fad961] text-[10px] font-medium px-2 py-0.5 rounded-md text-slate-800 shadow-sm z-20">
                          5h 23m
                        </div>
                      )}
                    </div>
                  )})}
                </div>
              </div>

              {/* Time Tracker */}
              <div className="flex-[0.45] bg-white/70 backdrop-blur-md rounded-[32px] p-6 shadow-sm relative group flex flex-col items-center">
                <div className="absolute top-5 right-5 p-1.5 rounded-full border border-slate-200 group-hover:bg-slate-50 transition-colors">
                  <ArrowUpRight size={14} className="text-slate-500" />
                </div>
                <h3 className="text-[17px] font-medium self-start w-full text-slate-800">Time tracker</h3>
                
                <div className="relative mt-2 mb-2 w-24 h-24 flex items-center justify-center">
                   <div className="absolute inset-0 rounded-full border-[6px] border-slate-100"></div>
                   <div className="absolute inset-0 rounded-full border-[6px] border-transparent border-r-[#fad961] border-b-[#fad961] transform -rotate-45 shadow-[0_0_15px_rgba(250,217,97,0.3)]"></div>
                   {/* dashed tick marks around */}
                   <div className="absolute inset-[-8px] border-[1px] border-dashed border-slate-300 rounded-full opacity-50"></div>
                   <div className="text-center z-10 relative mt-1">
                     <span className="text-[26px] font-light text-slate-800 tracking-tight leading-none">02:35</span>
                     <p className="text-[9px] text-slate-500 font-medium mt-0.5">Work Time</p>
                   </div>
                </div>

                <div className="flex items-center gap-2 mt-auto self-start pl-2">
                  <button className="w-8 h-8 flex items-center justify-center text-slate-400 hover:text-slate-800 transition-colors">
                    <Play size={15} fill="currentColor" />
                  </button>
                  <button className="w-8 h-8 flex items-center justify-center text-slate-800">
                    <Pause size={15} fill="currentColor" />
                  </button>
                  <div className="flex-1"></div>
                  <button className="w-9 h-9 rounded-[14px] bg-[#292b2a] text-white flex items-center justify-center hover:bg-black shadow-[0_4px_10px_rgba(41,43,42,0.3)] absolute right-5 bottom-5">
                    <Clock size={16} />
                  </button>
                </div>
              </div>
            </div>

            {/* Calendar */}
            <div className="bg-white/70 backdrop-blur-md rounded-[32px] p-6 pt-5 shadow-sm flex-1 flex flex-col relative overflow-hidden">
              {/* gradient blobs for BG lighting */}
              <div className="absolute -bottom-10 -right-10 w-40 h-40 bg-[#f4d89a]/30 blur-2xl rounded-full"></div>
              
              <div className="flex justify-between items-center mb-4 z-10">
                <button className="px-5 py-1 rounded-full border border-slate-200 text-[11px] font-medium bg-white text-slate-600 shadow-sm">August</button>
                <h3 className="font-medium text-slate-800 text-sm">September 2024</h3>
                <button className="px-5 py-1 rounded-full border border-slate-200 text-[11px] font-medium bg-white text-slate-600 shadow-sm">October</button>
              </div>

              <div className="grid grid-cols-6 gap-0 text-center text-xs text-slate-400 mb-1 z-10">
                <div></div>
                <div><span className="text-[10px] font-medium">Mon</span><br/><span className="text-slate-800 font-medium text-[13px] mt-1 inline-block">22</span></div>
                <div><span className="text-[10px] font-medium">Tue</span><br/><span className="text-slate-800 font-medium text-[13px] mt-1 inline-block">23</span></div>
                <div><span className="text-[10px] font-medium">Wed</span><br/><span className="text-slate-800 font-medium text-[13px] mt-1 inline-block">24</span></div>
                <div><span className="text-[10px] font-medium">Thu</span><br/><span className="text-slate-400 font-medium text-[13px] mt-1 inline-block">25</span></div>
                <div><span className="text-[10px] font-medium">Fri</span><br/><span className="text-slate-400 font-medium text-[13px] mt-1 inline-block">26</span></div>
              </div>

              <div className="relative flex-1 mt-1 z-10">
                <div className="absolute inset-0 grid grid-cols-6 gap-0 px-2 lg:px-4">
                  <div className="border-r border-slate-200/60 border-dashed text-[10px] font-medium text-slate-400 space-y-6 text-right pr-3 pt-3">
                    <div>8:00 am</div>
                    <div>9:00 am</div>
                    <div>10:00 am</div>
                    <div>11:00 am</div>
                  </div>
                  <div className="border-r border-slate-200/60 border-dashed relative">
                     {/* Blue block span */}
                  </div>
                  <div className="border-r border-slate-200/60 border-dashed relative">
                    <div className="absolute top-[28px] -left-[90px] w-[200px] h-[46px] bg-[#292b2a] rounded-[14px] flex items-center p-3 gap-2 shadow-[0_8px_15px_rgba(0,0,0,0.15)] z-20 text-white">
                      <div className="flex-1 overflow-hidden">
                        <p className="text-[11px] font-medium mb-0.5">Weekly Team Sync</p>
                        <p className="text-[9px] text-white/50 w-full truncate">Discuss progress on projects</p>
                      </div>
                      <div className="flex -space-x-1.5 shrink-0">
                        <img className="w-[18px] h-[18px] rounded-full border border-[#292b2a] object-cover" src="https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=100&h=100&fit=crop" />
                        <img className="w-[18px] h-[18px] rounded-full border border-[#292b2a] object-cover" src="https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=100&h=100&fit=crop" />
                        <img className="w-[18px] h-[18px] rounded-full border border-[#292b2a] object-cover" src="https://images.unsplash.com/photo-1531746020798-e6953c6e8e04?w=100&h=100&fit=crop" />
                      </div>
                    </div>
                  </div>
                  <div className="border-r border-slate-200/60 border-dashed relative">
                  </div>
                  <div className="border-r border-slate-200/60 border-dashed relative">
                    <div className="absolute top-[85px] -left-[70px] w-[180px] h-[46px] bg-white border border-slate-100 rounded-[14px] flex items-center p-3 gap-2 shadow-[0_8px_15px_rgba(0,0,0,0.05)] z-20">
                      <div className="flex-1 overflow-hidden">
                        <p className="text-[11px] font-medium text-slate-800 mb-0.5">Onboarding Session</p>
                        <p className="text-[9px] text-slate-400 w-full truncate">Introduction for new hires</p>
                      </div>
                      <div className="flex -space-x-1.5 shrink-0">
                        <img className="w-[18px] h-[18px] rounded-full border border-white object-cover" src="https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&h=100&fit=crop" />
                        <img className="w-[18px] h-[18px] rounded-full border border-white object-cover" src="https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=100&h=100&fit=crop" />
                      </div>
                    </div>
                  </div>
                  <div></div>
                </div>
              </div>
            </div>
          </div>

          {/* COL 3 */}
          <div className="col-span-3 flex flex-col gap-5">
            <div className="bg-white/70 backdrop-blur-md rounded-[32px] p-6 shadow-sm h-[130px]">
              <div className="flex justify-between items-start mb-4">
                <h3 className="text-[17px] font-medium text-slate-800">Onboarding</h3>
                <span className="text-3xl font-light leading-none -mt-1">18%</span>
              </div>
              <div className="flex items-center text-[10px] text-slate-500 mb-1.5 gap-1">
                <div className="flex-[0.4] font-medium">30%</div>
                <div className="flex-[0.4] font-medium">25%</div>
                <div className="flex-[0.2] font-medium">0%</div>
              </div>
              <div className="flex gap-1 h-7 w-full">
                <div className="flex-[0.4] bg-[#fad961] rounded-l-full rounded-r-md flex items-center px-3 text-[10px] font-medium text-slate-800 shadow-[0_2px_8px_rgba(250,217,97,0.3)]">Task</div>
                <div className="flex-[0.4] bg-[#292b2a] rounded-md shadow-[0_2px_8px_rgba(0,0,0,0.1)]"></div>
                <div className="flex-[0.2] bg-slate-300/60 rounded-r-full rounded-l-md"></div>
              </div>
            </div>

            <div className="bg-[#292b2a] rounded-[32px] p-7 pt-6 text-white flex-1 shadow-[0_20px_40px_rgba(0,0,0,0.12)] flex flex-col">
              <div className="flex justify-between items-start mb-6">
                <h3 className="text-[17px] font-medium text-white/90 leading-tight">Onboarding Task</h3>
                <span className="text-3xl font-light leading-none text-white/90">2/8</span>
              </div>

              <div className="space-y-5 mt-1">
                {[
                  { icon: Monitor, title: 'interview', time: 'Sep 13, 08:30', done: true },
                  { icon: Zap, title: 'Team-Meeting', time: 'Sep 13, 10:30', done: true },
                  { icon: MessageSquare, title: 'Project Update', time: 'Sep 13, 13:00', done: false },
                  { icon: Briefcase, title: 'Discuss Q3 Goals', time: 'Sep 13, 14:45', done: false },
                  { icon: LinkIcon, title: 'HR Policy Review', time: 'Sep 13, 16:30', done: false },
                ].map((task, i) => (
                  <div key={i} className="flex items-center gap-3.5 group cursor-pointer">
                    <div className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 transition-colors ${task.done ? 'bg-white/10' : 'bg-white/5 group-hover:bg-white/10'}`}>
                       <task.icon size={14} className={task.done ? "text-white/90" : "text-white/40"} />
                    </div>
                    <div className="flex-1">
                      <p className={`text-[13px] ${task.done ? 'text-white/90' : 'text-white/50'} font-medium leading-tight`}>{task.title}</p>
                      <p className="text-[10px] text-white/30 mt-0.5">{task.time}</p>
                    </div>
                    <div>
                      {task.done ? (
                        <div className="w-5 h-5 rounded-full bg-[#fad961] flex items-center justify-center shadow-[0_0_8px_rgba(250,217,97,0.3)]">
                          <CheckCircle2 size={13} className="text-[#292b2a]" strokeWidth={3} />
                        </div>
                      ) : (
                        <div className="w-5 h-5 rounded-full border-[1.5px] border-white/20 group-hover:border-white/40 transition-colors"></div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>
  );
}
