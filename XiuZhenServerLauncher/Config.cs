using System.Collections.Generic;

namespace XiuZhenServerLauncher
{
    class Config
    {
        public Ftp Ftp { get; set; }
        public List<string> Whitelist { get; set; }
        public List<Bat> Bats { get; set; }
    }

    class Bat
    {
        public int RegionId { get; set; }
        public string WorkDir { get; set; }
        public string Start { get; set; }
        public string Stop { get; set; }
    }

    class Ftp
    {
        public string Host { get; set; }
        public string User { get; set; }
        public string Password { get; set; }
    }
}
