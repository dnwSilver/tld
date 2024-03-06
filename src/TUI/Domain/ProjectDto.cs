using System.Runtime.Serialization;
using TUI.Settings;
using YamlDotNet.Serialization;


namespace TUI.Domain;

[DataContract]
[YamlSerializable]
public class ProjectDto
{
    [YamlMember]
    [DataMember]
    public string Icon { get; set; }

    [YamlMember]
    [DataMember]
    public string Name { get; set; }

    [YamlMember]
    [DataMember]
    public DependencyDto[] Dependencies { get; set; }

    [YamlMember]
    [DataMember]
    public IList<SourceDto> Sources { get; set; }
}