%Image::ExifTool::UserDefined = (
    'Image::ExifTool::XMP::crs' => {
        Look => {
            Struct => {
                STRUCT_NAME => 'Look',
                NAMESPACE   => 'crs',
                Name   => { },
                Amount => { },
                UUID   => { },
                Parameters => {
                    Struct => {
                        STRUCT_NAME => 'Parameters',
                        NAMESPACE   => 'crs',
                        ProcessVersion           => { },
                        CameraProfile            => { },
                        ConvertToGrayscale       => { Writable => 'boolean' },
                        Exposure2012             => { Writable => 'real' },
                        Contrast2012             => { Writable => 'integer' },
                        Highlights2012           => { Writable => 'integer' },
                        Shadows2012              => { Writable => 'integer' },
                        Whites2012               => { Writable => 'integer' },
                        Blacks2012               => { Writable => 'integer' },
                        Clarity2012              => { Writable => 'integer' },
                        IncrementalTemperature   => { Writable => 'integer' },
                        IncrementalTint          => { Writable => 'integer' },
                        ParametricShadows        => { Writable => 'integer' },
                        ParametricDarks          => { Writable => 'integer' },
                        ParametricLights         => { Writable => 'integer' },
                        ParametricHighlights     => { Writable => 'integer' },
                        ParametricShadowSplit    => { Writable => 'integer' },
                        ParametricMidtoneSplit   => { Writable => 'integer' },
                        ParametricHighlightSplit => { Writable => 'integer' },
                        ToneCurveName2012        => { },
                        ToneCurvePV2012          => { List => 'Seq' },
                        ToneCurvePV2012Red       => { List => 'Seq' },
                        ToneCurvePV2012Green     => { List => 'Seq' },
                        ToneCurvePV2012Blue      => { List => 'Seq' },
                        LookTable                => { },
                    }
                }
            }
        }
    }
)