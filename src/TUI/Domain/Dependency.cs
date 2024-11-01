using TUI.Controls.Components;

namespace TUI.Domain;

public record Dependency()
{
    private readonly Version _current;
    public readonly Brand Brand;
    public VersionType Type { get; private set; }
    
    public string Version => $"{_current.Major}.{_current.Minor}.{_current.Patch}";
    
    public Dependency(string version, Brand brand) : this()
    {
        _current = new Version(version);
        Brand = brand;
    }
    
    public VersionStatus Comparison(Dependency outerDependency)
    {
        // if (string.IsNullOrEmpty(Version) || string.IsNullOrEmpty(outerDependency.Version))
        // {
        // return VersionStatus.NotFound;
        // }
        
        var outer = outerDependency._current;
        
        Type = _current.Type;
        
        if (_current.Major < outer.Major)
        {
            return VersionStatus.TooOld;
        }
        
        if (_current.Major > outer.Major)
        {
            return VersionStatus.ToNew;
        }
        
        if (_current.Minor < outer.Minor)
        {
            return VersionStatus.BeNice;
        }
        
        if (_current.Minor > outer.Minor)
        {
            return VersionStatus.ToNew;
        }
        
        if (_current.Patch < outer.Patch)
        {
            return VersionStatus.BeNice;
        }
        
        if (_current.Patch > outer.Patch)
        {
            return VersionStatus.ToNew;
        }
        
        if (outer.Type != VersionType.Release)
        {
            return VersionStatus.ToNew;
        }
        
        return VersionStatus.SoGood;
    }
};